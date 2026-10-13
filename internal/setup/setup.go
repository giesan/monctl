// Package setup configures the Linux i2c-dev module for DDC/CI access.
package setup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const configuration = "i2c-dev\n"

type runner interface {
	Run(context.Context, io.Reader, io.Writer, io.Writer, string, ...string) error
}

type executable struct{}

func (executable) Run(ctx context.Context, in io.Reader, out, errOut io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = in, out, errOut
	return cmd.Run()
}

type Setup struct {
	configPath string
	modulePath string
	executable string
	euid       func() int
	run        runner
	in         io.Reader
}

func New() (*Setup, error) {
	path, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("locate monctl executable: %w", err)
	}
	return &Setup{
		configPath: "/etc/modules-load.d/i2c-dev.conf",
		modulePath: "/sys/module/i2c_dev",
		executable: path,
		euid:       os.Geteuid,
		run:        executable{},
		in:         os.Stdin,
	}, nil
}

func (s *Setup) status() (loaded, configured bool, err error) {
	_, err = os.Stat(s.modulePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, false, fmt.Errorf("check i2c-dev module: %w", err)
	}
	loaded = err == nil
	contents, err := os.ReadFile(s.configPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, false, fmt.Errorf("read module configuration: %w", err)
	}
	if err == nil && strings.TrimSpace(string(contents)) != "i2c-dev" {
		return false, false, fmt.Errorf("%s already exists with different contents; inspect it before running setup", s.configPath)
	}
	return loaded, err == nil, nil
}

// Run asks for privileges only when a system change is needed.
func (s *Setup) Run(ctx context.Context, out, errOut io.Writer) error {
	loaded, configured, err := s.status()
	if err != nil {
		return err
	}
	if loaded && configured {
		fmt.Fprintln(out, "i2c-dev is loaded and configured to start at boot.")
		return nil
	}
	if !loaded {
		fmt.Fprintln(out, "Will load the i2c-dev kernel module.")
	}
	if !configured {
		fmt.Fprintf(out, "Will create %s to load it at boot.\n", s.configPath)
	}
	if s.euid() == 0 {
		err = s.Apply(ctx)
	} else {
		err = s.run.Run(ctx, s.in, out, errOut, "sudo", "--", s.executable, "setup", "--apply")
	}
	if err != nil {
		return fmt.Errorf("configure i2c-dev: %w", err)
	}
	fmt.Fprintln(out, "i2c-dev setup complete.")
	return nil
}

// Apply is the privileged, narrowly scoped part of setup.
func (s *Setup) Apply(ctx context.Context) error {
	if s.euid() != 0 {
		return errors.New("setup --apply requires administrator privileges")
	}
	loaded, configured, err := s.status()
	if err != nil {
		return err
	}
	if !loaded {
		var diagnostic bytes.Buffer
		if err := s.run.Run(ctx, s.in, io.Discard, &diagnostic, "modprobe", "i2c-dev"); err != nil {
			return fmt.Errorf("load i2c-dev with modprobe: %s: %w", strings.TrimSpace(diagnostic.String()), err)
		}
	}
	if !configured {
		file, err := os.OpenFile(s.configPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if errors.Is(err, os.ErrExist) {
			// Another setup process may have created the file in the meantime.
			_, configured, err = s.status()
			if err != nil {
				return err
			}
			if configured {
				return nil
			}
		}
		if err != nil {
			return fmt.Errorf("create %s: %w", s.configPath, err)
		}
		_, writeErr := io.WriteString(file, configuration)
		closeErr := file.Close()
		if writeErr != nil {
			return fmt.Errorf("write %s: %w", s.configPath, writeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close %s: %w", s.configPath, closeErr)
		}
	}
	return nil
}

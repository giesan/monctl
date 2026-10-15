# Ein Fenster statt Tray — Entwurf

## Ziel und Verhalten

`monctl` ist das einzige ausgelieferte Programm. Ohne Argumente öffnet es das vorhandene Helligkeitsfenster mit Reglern für erkannte externe Monitore und, falls vorhanden, das Laptop-Display. Das Schließen des Fensters beendet das Programm. Die bisherigen CLI-Befehle (`setup`, `list`, `get`, `set`, `contrast`, `volume`, `laptop`) bleiben mit ihren Argumenten und Rückgabewerten erhalten und funktionieren ohne grafische Sitzung. Unbekannte Argumente zeigen weiterhin die Hilfe mit Exit-Code 2.

Das Fenster nutzt weiterhin Nuculars CGO-freies Shiny-Backend und benötigt X11 oder XWayland. Fehlt `DISPLAY`, meldet `monctl` den Fehler und beendet sich mit Exit-Code 1. Wayland ohne XWayland ist nicht Teil dieses Schritts.

## Aufbau

Der CLI-Einstieg unterscheidet lediglich zwischen Aufruf ohne Argumente und einem CLI-Befehl. Der Fensterpfad ruft `internal/window.Run` mit den vorhandenen Monitor- und Laptop-Diensten auf; das Fenster-Modell und seine verzögerten Schreibzugriffe bleiben bestehen. Die GUI-spezifische Anbindung im CLI ist an den Build-Tag `nucular_shiny` gebunden. Ohne diesen Tag bleibt die CLI baubar; ein Aufruf ohne Argumente gibt eine verständliche Build-Anweisung aus. Das ausgelieferte `monctl` wird mit dem Tag und `CGO_ENABLED=0` gebaut. Ein Fensterfehler wird auf stderr ausgegeben und führt zu Exit-Code 1.

Tray-Einstieg, Tray-Paket und der separate `monctl-window`-Einstieg werden entfernt. Die Tray-spezifische Implementierung und ihre Tests entfallen; `internal/window` bleibt als eigener Baustein bestehen. Go-Abhängigkeiten, die danach nicht mehr benötigt werden, werden bereinigt. Die Paketkonfiguration liefert nur `monctl` aus. Das README beschreibt den Start des Fensters, den Quellcode-Aufruf mit Build-Tag und die weiterhin verfügbaren CLI-Befehle.

## Prüfung

Tests prüfen die Verzweigung ohne Argumente einschließlich Fensterfehler und die unveränderte CLI-Argumentbehandlung, ohne ein Display oder echte Monitorhardware vorauszusetzen. `go test ./...` und `go test -tags=nucular_shiny ./...` prüfen beide Build-Varianten; ein Paket-Build prüft die Auslieferung mit nur einem Binary. Eine grafische Sichtprüfung auf X11/XWayland bestätigt Öffnen, Regler und Beenden beim Schließen.

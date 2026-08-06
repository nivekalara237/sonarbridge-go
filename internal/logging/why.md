[!NOTE]

* **Server** → slog.NewJSONHandler(os.Stdout)
* **CLI** → slog.NewTextHandler(os.Stderr)
* **Résultat CLI** → stdout
* **Logs CLI** → stderr 
* **Un seul package** internal/logging et un logger global.
# Required Airflow Variables for Transit DAGs

Set the following in Airflow UI (Admin → Variables) or via CLI before running transit DAGs.
Contact the project lead for actual credential values.

| Key | Value |
|-----|-------|
| TDX_CLIENT_ID | `<your-tdx-client-id>` |
| TDX_CLIENT_SECRET | `<your-tdx-client-secret>` |

CLI:
```bash
airflow variables set TDX_CLIENT_ID "<your-tdx-client-id>"
airflow variables set TDX_CLIENT_SECRET "<your-tdx-client-secret>"
```

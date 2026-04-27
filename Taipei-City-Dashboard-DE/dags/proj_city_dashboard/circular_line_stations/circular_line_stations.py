from operators.common_pipeline import CommonDag


def _circular_line_stations(**kwargs):
    import pandas as pd
    from sqlalchemy import create_engine, text
    from utils.extract_stage import get_tdx_data

    # Config
    ready_data_db_uri = kwargs.get("ready_data_db_uri")
    dag_infos = kwargs.get("dag_infos")
    dag_id = dag_infos.get("dag_id")
    default_table = dag_infos.get("ready_data_default_table")

    SYSTEM = "ntmetro"
    URL = "https://tdx.transportdata.tw/api/basic/v2/Rail/Metro/Station/NTMETRO?$format=JSON"

    # Extract
    raw_data = get_tdx_data(URL, output_format="dataframe")

    # Transform
    data = raw_data.copy()
    data["system"] = SYSTEM
    data["station_id"] = data["StationID"].astype(str)
    data["station_name_zh"] = data["StationName"].apply(
        lambda x: x.get("Zh_tw", "") if isinstance(x, dict) else ""
    )
    data["station_name_en"] = data["StationName"].apply(
        lambda x: x.get("En", "") if isinstance(x, dict) else ""
    )
    data["lat"] = data["StationPosition"].apply(
        lambda x: x.get("PositionLat", None) if isinstance(x, dict) else None
    )
    data["lng"] = data["StationPosition"].apply(
        lambda x: x.get("PositionLon", None) if isinstance(x, dict) else None
    )
    # LineID is present in TDX NTMETRO response; fallback "YL" = Yellow Line (Circle Line)
    data["line_id"] = data["LineID"] if "LineID" in data.columns else "YL"
    data["sequence"] = data["StationSequence"] if "StationSequence" in data.columns else pd.RangeIndex(len(data))
    data["data_time"] = pd.Timestamp.now(tz="Asia/Taipei")

    ready_data = data[[
        "system", "line_id", "station_id",
        "station_name_zh", "station_name_en",
        "lat", "lng", "sequence", "data_time"
    ]]

    # Load — replace only ntmetro rows atomically
    engine = create_engine(ready_data_db_uri)
    with engine.begin() as conn:
        conn.execute(
            text("DELETE FROM transit_stations WHERE system = :sys"),
            {"sys": SYSTEM}
        )
        ready_data.to_sql(
            default_table, conn, if_exists="append", index=False, schema="public"
        )

    # Update dataset last-updated metadata
    from utils.load_stage import update_lasttime_in_data_to_dataset_info
    update_lasttime_in_data_to_dataset_info(
        engine, airflow_dag_id=dag_id, lasttime_in_data=ready_data["data_time"].max()
    )


dag = CommonDag(
    proj_folder="proj_city_dashboard", dag_folder="circular_line_stations"
)
dag.create_dag(etl_func=_circular_line_stations)

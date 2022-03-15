from datetime import datetime, timedelta
import json
# The DAG object; we'll need this to instantiate a DAG
from airflow import DAG

# Operators; we need this to operate!
from airflow.operators.bash import BashOperator
from airflow.operators.python import PythonOperator
# These args will get passed on to each operator
# You can override them on a per-task basis during operator initialization
default_args = {
    'owner': 'airflow',
    'depends_on_past': False,
    'email': ['airflow@example.com'],
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
    # 'queue': 'bash_queue',
    # 'pool': 'backfill',
    # 'priority_weight': 10,
    # 'end_date': datetime(2016, 1, 1),
    # 'wait_for_downstream': False,
    # 'dag': dag,
    # 'sla': timedelta(hours=2),
    # 'execution_timeout': timedelta(seconds=300),
    # 'on_failure_callback': some_function,
    # 'on_success_callback': some_other_function,
    # 'on_retry_callback': another_function,
    # 'sla_miss_callback': yet_another_function,
    # 'trigger_rule': 'all_success'
}
with DAG(
        '0_gh_passing_values',
        default_args=default_args,
        description='example_of_passing_values',
        schedule_interval=timedelta(days=1),
        start_date=datetime(2021, 1, 1),
        catchup=False,
        tags=['example666'],
        render_template_as_native_obj=True
) as dag:

    def extract():
        data_string = '{"1001": 301.27, "1002": 433.21, "1003": 502.22}'
        return json.loads(data_string)


    def transform(order_data):
        print(type(order_data))
        print(order_data)
        total_order_value = 0
        for value in order_data.values():
            total_order_value += value
        return {"total_order_value": total_order_value}


    extract_task = PythonOperator(
        task_id="extract",
        python_callable=extract,
    )

    transform_task = PythonOperator(
        task_id="transform",
        op_kwargs={"order_data": "{{ti.xcom_pull('extract')}}"},
        python_callable=transform,
    )

    extract_task >> transform_task



if __name__ == "__main__":
    dag.cli()
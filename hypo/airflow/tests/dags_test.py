import os

from airflow.models import DagBag


def test_dag_imports():
    dag_bag = DagBag(include_examples=False, dag_folder=os.getcwd() + "/..")
    print(dag_bag.import_errors)
    assert len(dag_bag.import_errors) == 0, "Import Failures"
import os
import unittest

from airflow.models import DagBag
from dags.gh_dag.e1 import sum33


class TestSum(unittest.TestCase):
    def test_list_int(self):
        """
        Test that it can sum a list of integers
        """
        data = [1, 2, 3]
        result = sum33(data)
        self.assertEqual(result, 6)

# if __name__ == '__main__':
#     unittest.main()

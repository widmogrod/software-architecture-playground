from pyspark import SparkContext
from pyspark.streaming import StreamingContext

if __name__ == "__main__":
    # Create a local StreamingContext with two working thread and batch interval of 1 second
    sc = SparkContext("local[2]", "Parallel")

    nums = sc.parallelize([1, 2, 3, 4, 5, 6])
    res = nums.fold(0, lambda a, b: a + b)
    # nums = nums.fold(lambda i: print(i))
    nums.collect()

    print(res)

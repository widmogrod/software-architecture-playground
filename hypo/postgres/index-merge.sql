drop table if exists index_merge;
create table if not exists index_merge
(
    content int,
    market int
);

insert into index_merge(content, market)
SELECT random() * 10000, random() * 100000 FROM generate_series(0,1000000);

create index idx_content on index_merge(content);
create index idx_market on index_merge(market);

explain (analyze, buffers) select content,market from index_merge where content = 1000 and market = 9090;
-- Bitmap Heap Scan on index_merge  (cost=188.11..282.85 rows=25 width=8) (actual time=0.219..0.281 rows=0 loops=1)
--   Recheck Cond: ((market = 9090) AND (content = 1000))
--   Buffers: shared read=6
--   ->  BitmapAnd  (cost=188.11..188.11 rows=25 width=0) (actual time=0.177..0.221 rows=0 loops=1)
--         Buffers: shared read=6
--         ->  Bitmap Index Scan on idx_market  (cost=0.00..93.92 rows=5000 width=0) (actual time=0.051..0.060 rows=9 loops=1)
--               Index Cond: (market = 9090)
--               Buffers: shared read=3
--         ->  Bitmap Index Scan on idx_content  (cost=0.00..93.92 rows=5000 width=0) (actual time=0.051..0.060 rows=100 loops=1)
--               Index Cond: (content = 1000)
--               Buffers: shared read=3
-- Planning Time: 0.360 ms
-- Execution Time: 0.396 ms


explain (analyze, buffers) select content,market from index_merge where content = 1000 or market = 9090;
-- Bitmap Heap Scan on index_merge  (cost=9.74..399.37 rows=111 width=8) (actual time=0.180..1.735 rows=112 loops=1)
--   Recheck Cond: ((content = 1000) OR (market = 9090))
--   Heap Blocks: exact=111
--   Buffers: shared hit=117
--   ->  BitmapOr  (cost=9.74..9.74 rows=111 width=0) (actual time=0.100..0.141 rows=0 loops=1)
--         Buffers: shared hit=6
--         ->  Bitmap Index Scan on idx_content  (cost=0.00..5.17 rows=100 width=0) (actual time=0.048..0.057 rows=99 loops=1)
--               Index Cond: (content = 1000)
--               Buffers: shared hit=3
--         ->  Bitmap Index Scan on idx_market  (cost=0.00..4.51 rows=11 width=0) (actual time=0.024..0.031 rows=13 loops=1)
--               Index Cond: (market = 9090)
--               Buffers: shared hit=3
-- Planning Time: 0.092 ms
-- Execution Time: 3.108 ms
drop table if exists index_merge;
create table if not exists index_merge
(
    content int,
    market int
);

insert into index_merge(content, market)
SELECT random() * 10000, random() * 100000 FROM generate_series(0,1000000);

create index idx_content_market on index_merge(content, market);

explain (analyze, buffers) select content,market from index_merge where content = 1000 and market = 9090;
-- Index Only Scan using idx_content_market on index_merge  (cost=0.42..8.45 rows=1 width=8) (actual time=0.029..0.038 rows=0 loops=1)
--   Index Cond: ((content = 1000) AND (market = 9090))
--   Heap Fetches: 0
--   Buffers: shared hit=3
-- Planning Time: 0.109 ms
-- Execution Time: 0.102 ms


explain (analyze, buffers) select content,market from index_merge where content = 1000 or market = 9090;
-- Gather  (cost=1000.00..11686.11 rows=111 width=8) (actual time=2.540..85.909 rows=127 loops=1)
--   Workers Planned: 2
--   Workers Launched: 2
--   Buffers: shared hit=4425
--   ->  Parallel Seq Scan on index_merge  (cost=0.00..10675.01 rows=46 width=8) (actual time=1.794..55.067 rows=42 loops=3)
--         Filter: ((content = 1000) OR (market = 9090))
--         Rows Removed by Filter: 333291
--         Buffers: shared hit=4425
-- Planning Time: 0.462 ms
-- Execution Time: 87.839 ms

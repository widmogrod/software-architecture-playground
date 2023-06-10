## TODO
## v0.1.0
- [ ] In memory interpretation with error handling (not panics) and stoping DAGs
- [ ] Kinesis Stream with simplere multi subscriber implementation
- [ ] Deploy on separate Stack different architecture concepts (OpenSearch, Synchronous, Live Select games)

## v0.2.0
- [ ] In memory interpretation with graceful shutdown
- [ ] Support scalability of each dag node
- [ ] Introduce concept of control plain, so that each dag can assume roles
- [ ] In memory interpretation is control plain aware


Configuration of projection runtime could looke like this:
```yaml
default:
  instances: 1
  resources:
    cpu: 0.5
    memory: 100Mi
    
dag:
  - name: "DynamoDB Load Stream"
    config:
      autoscaling: fixed
      fixed:
          instances: 2
      resources: 
        cpu: 1
        memory: 100Mi
```

This is how control plane and scaling of projection layer could help 
DAG run, and allocate resources to what is needed. There is also option to addd autoscaking

Having such state, also suggest that w binnary of bigger DAGs, can be deployed as one on cluster
and sub-dags could have dedicated resources. In example, when DAG starts with stream of events,
but there are many processes that follow from that, but branch into thousand different processes. 

Thousand is naturaly,exaggeration, but it is to create though exercise how system can evolve.
Which means that new sub-dags, can be faulty, how to ensure that new deployment, won't cause blast radius to 
already well established processes?

What if thouse thousand sub-dags, work perfectly fine, but they require 1000 servers to run on?
Should change of one sub-dag, cause all other sub-dags to be restarted?
Should only a sub-dag be submitted to be redeployed and restarted?

I believe that there is no good answer, and that both options are valid.
Which means, that the solution add possibility to deploy sub-dag on the same or maybe separate cluster.
Which means that sub-dag needs to connect to stream of operations from different cluster.
Which means that communication between nodes, between different cluster, should be referenced, 
and this also introduce interesting problem, what if a schema changes on producing node, and consiming node is not changed?

This also means, that concept of schemas bounded to nodes, is important, to validate formats.
And consuming node should stop, when types wont match. Deploying monolithic dag, 
can make sure that such bugs will be caught in compile time, even if deployment will be to separate clustes.

Versioning of stream, would help in such a way, that consimign node that use older version of schema, would stop recieveing messages
And it would require explicit action to switch to newer model. When a stream that was deprecated will be closed, then consiming node will stop recieving messages, and start to fail.

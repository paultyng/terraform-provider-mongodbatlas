# Data Source: mongodbatlas_stream_processor

`mongodbatlas_stream_processor` describes a stream processor.

## Example Usages
```terraform
resource "mongodbatlas_stream_instance" "example" {
  project_id    = var.project_id
  instance_name = "InstanceName"
  data_process_region = {
    region         = "VIRGINIA_USA"
    cloud_provider = "AWS"
  }
}

resource "mongodbatlas_stream_connection" "example-sample" {
  project_id      = var.project_id
  instance_name   = mongodbatlas_stream_instance.example.instance_name
  connection_name = "sample_stream_solar"
  type            = "Sample"
}

resource "mongodbatlas_stream_connection" "example-cluster" {
  project_id      = var.project_id
  instance_name   = mongodbatlas_stream_instance.example.instance_name
  connection_name = "ClusterConnection"
  type            = "Cluster"
  cluster_name    = var.cluster_name
  db_role_to_execute = {
    role = "atlasAdmin"
    type = "BUILT_IN"
  }
}

resource "mongodbatlas_stream_connection" "example-kafka" {
  project_id      = var.project_id
  instance_name   = mongodbatlas_stream_instance.example.instance_name
  connection_name = "KafkaPlaintextConnection"
  type            = "Kafka"
  authentication = {
    mechanism = "PLAIN"
    username  = var.kafka_username
    password  = var.kafka_password
  }
  bootstrap_servers = "localhost:9092,localhost:9092"
  config = {
    "auto.offset.reset" : "earliest"
  }
  security = {
    protocol = "PLAINTEXT"
  }
}

resource "mongodbatlas_stream_processor" "stream-processor-sample-example" {
  project_id     = var.project_id
  instance_name  = mongodbatlas_stream_instance.example.instance_name
  processor_name = "sampleProcessorName"
  pipeline       = jsonencode([{ "$source" = { "connectionName" = resource.mongodbatlas_stream_connection.example-sample.connection_name } }, { "$emit" = { "connectionName" : "__testLog" } }])
  state          = "CREATED"
}

resource "mongodbatlas_stream_processor" "stream-processor-cluster-example" {
  project_id     = var.project_id
  instance_name  = mongodbatlas_stream_instance.example.instance_name
  processor_name = "clusterProcessorName"
  pipeline       = jsonencode([{ "$source" = { "connectionName" = resource.mongodbatlas_stream_connection.example-cluster.connection_name } }, { "$emit" = { "connectionName" : "__testLog" } }])
  state          = "STARTED"
}

resource "mongodbatlas_stream_processor" "stream-processor-kafka-example" {
  project_id     = var.project_id
  instance_name  = mongodbatlas_stream_instance.example.instance_name
  processor_name = "kafkaProcessorName"
  pipeline       = jsonencode([{ "$source" = { "connectionName" = resource.mongodbatlas_stream_connection.example-cluster.connection_name } }, { "$emit" = { "connectionName" : resource.mongodbatlas_stream_connection.example-kafka.connection_name, "topic" : "example_topic" } }])
  state          = "CREATED"
  options = {
    dlq = {
      coll            = "exampleColumn"
      connection_name = resource.mongodbatlas_stream_connection.example-cluster.connection_name
      db              = "exampleDb"
    }
  }
}

data "mongodbatlas_stream_processors" "example-stream-processors" {
  project_id    = var.project_id
  instance_name = mongodbatlas_stream_instance.example.instance_name
}

data "mongodbatlas_stream_processor" "example-stream-processor" {
  project_id     = var.project_id
  instance_name  = mongodbatlas_stream_instance.example.instance_name
  processor_name = mongodbatlas_stream_processor.stream-processor-sample-example.processor_name
}

# example making use of data sources
output "stream_processors_state" {
  value = data.mongodbatlas_stream_processor.example-stream-processor.state
}

output "stream_processors_results" {
  value = data.mongodbatlas_stream_processors.example-stream-processors.results
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `instance_name` (String) Human-readable label that identifies the stream instance.
- `processor_name` (String) Human-readable label that identifies the stream processor.
- `project_id` (String) Unique 24-hexadecimal digit string that identifies your project. Use the [/groups](#tag/Projects/operation/listProjects) endpoint to retrieve all projects to which the authenticated user has access.

**NOTE**: Groups and projects are synonymous terms. Your group id is the same as your project id. For existing groups, your group/project id remains the same. The resource and corresponding endpoints use the term groups.

### Read-Only

- `id` (String) Unique 24-hexadecimal character string that identifies the stream processor.
- `options` (Attributes) Optional configuration for the stream processor. (see [below for nested schema](#nestedatt--options))
- `pipeline` (String) Stream aggregation pipeline you want to apply to your streaming data.
- `state` (String) The state of the stream processor.
- `stats` (String) The stats associated with the stream processor.

<a id="nestedatt--options"></a>
### Nested Schema for `options`

Read-Only:

- `dlq` (Attributes) Dead letter queue for the stream processor. (see [below for nested schema](#nestedatt--options--dlq))

<a id="nestedatt--options--dlq"></a>
### Nested Schema for `options.dlq`

Read-Only:

- `coll` (String) Name of the collection that will be used for the DLQ.
- `connection_name` (String) Connection name that will be used to write DLQ messages to. Has to be an Atlas connection.
- `db` (String) Name of the database that will be used for the DLQ.

For more information see: [MongoDB Atlas API - Stream Processor](https://www.mongodb.com/docs/atlas/reference/api-resources-spec/v2/#tag/Streams/operation/createStreamProcessor) Documentation.
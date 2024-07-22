using Confluent.Kafka;
using Grpc.Net.Client;
using System.Runtime.InteropServices;
using Microsoft.Extensions.Logging;
using LoadBalancer.Services;
using LoadBalancer.Kafka;

await Task.Delay(10000);

var postServiceAdresses = new List<string>() { "http://post-service1:8080", "http://post-service2:8080" };

var bootstrapServers = "kafka:9092";
var creationTopic = "create_post";
var groupId = "post_service_consumer_group";

var client = new PostServiceClient(postServiceAdresses);
var creationTopicConsumer = new KafkaConsumer(bootstrapServers, creationTopic, groupId, client);

var consumeTask = creationTopicConsumer.ConsumeAsync();

await consumeTask;

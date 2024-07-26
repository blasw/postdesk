using Confluent.Kafka;
using Grpc.Net.Client;
using System.Runtime.InteropServices;
using Microsoft.Extensions.Logging;
using LoadBalancer.Services;
using LoadBalancer.Kafka;
using Microsoft.Extensions.Configuration;

await Task.Delay(10000);

var configuration = new ConfigurationBuilder()
    .AddEnvironmentVariables()
    .Build();

var postServiceAdresses = configuration["POST_SERVICE_ADDRESSES"].Split(',').ToList();  

var bootstrapServers = configuration["KAFKA_BROKER"];
var creationTopic = configuration["CREATE_TOPIC"];
var groupId = configuration["GROUP"];


var client = new PostServiceClient(postServiceAdresses);
var creationTopicConsumer = new KafkaConsumer(bootstrapServers, creationTopic, groupId, client);

var consumeTask = creationTopicConsumer.ConsumeAsync();

await consumeTask;

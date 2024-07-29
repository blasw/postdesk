using LoadBalancer.Services;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;

namespace LoadBalancer.Kafka
{
    internal class KafkaConsumerFactory
    {
        private readonly IServiceProvider _serviceProvider;
        private readonly ILogger<KafkaConsumerFactory> _logger;

        public KafkaConsumerFactory(IServiceProvider serviceProvider, ILogger<KafkaConsumerFactory> logger)
        {
            _serviceProvider = serviceProvider;
            _logger = logger;
        }

        public KafkaConsumer CreateConsumer(string boostrapServers, string topic, string groupId)
        {
            var postServiceClient = _serviceProvider.GetRequiredService<PostServiceClient>();
            var logger = _serviceProvider.GetRequiredService<ILogger<KafkaConsumer>>();

            _logger.LogInformation($"Creating KafkaConsumer for topic: {topic}, groupId: {groupId}");

            return new KafkaConsumer(boostrapServers, topic, groupId, postServiceClient, logger, _serviceProvider);
        }
    }
}

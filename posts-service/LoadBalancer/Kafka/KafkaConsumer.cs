using LoadBalancer.Services;
using Confluent.Kafka;
using Microsoft.Extensions.Logging;
using LoadBalancer.Core;
using Microsoft.Extensions.DependencyInjection;

namespace LoadBalancer.Kafka
{
    public class KafkaConsumer
    {
        private readonly string _bootstrapServers;
        private readonly string _topic;
        private readonly string _groupId;

        private readonly PostServiceClient _postServiceClient;
        private readonly IServiceProvider _serviceProvider;
        private readonly ILogger<KafkaConsumer> _logger;

        public KafkaConsumer(string bootstrapServers, string topic, string groupId, PostServiceClient postServiceClient, ILogger<KafkaConsumer> logger, IServiceProvider serviceProvider)
        {
            _bootstrapServers = bootstrapServers;
            _topic = topic;
            _groupId = groupId;
            _postServiceClient = postServiceClient;
            _logger = logger;
            _serviceProvider = serviceProvider;
        }

        public async Task ConsumeAsync()
        {
            var config = new ConsumerConfig
            {
                GroupId = _groupId,
                BootstrapServers = _bootstrapServers,
                AutoOffsetReset = AutoOffsetReset.Earliest
            };

            using var consumer = new ConsumerBuilder<Ignore, string>(config).Build();

            while (true)
            {
                try
                {
                    consumer.Subscribe(_topic);

                    _logger.LogInformation("Subscribed to topic {Topic}.", _topic);

                    await ConsumeMessagesAsync(consumer);
                }
                catch (ConsumeException ex)
                {
                    _logger.LogError("Consume error: {Reason}", ex.Error.Reason);
                    if (ex.Error.Code == ErrorCode.UnknownTopicOrPart)
                    {
                        _logger.LogWarning("Topic '{Name}' does not exist. Waiting for it to become available...", _topic);
                    }
                    await Task.Delay(5000);
                }
                catch (Exception ex)
                {
                    _logger.LogError("Unexpected error: {Message}", ex.Message);
                    await Task.Delay(5000);
                }
                finally
                {
                    consumer.Unsubscribe();
                }
            }
        }

        private async Task ConsumeMessagesAsync(IConsumer<Ignore, string> consumer)
        {
            while (true)
            {
                try
                {
                    var consumeResult = consumer.Consume();
                    _logger.LogInformation("Consumed message '{Message}' at: '{Topic}'.", consumeResult.Message.Value, consumeResult.TopicPartitionOffset);

                    var handler = GetHandlerForTopic(_topic);
                    if (handler != null)
                    {
                        await handler.HandleMessageAsync(consumeResult.Message.Value);
                    }
                    else
                    {
                        _logger.LogWarning("No handler found for topic: {Topic}", _topic);
                    }
                }
                catch (ConsumeException ex)
                {
                    _logger.LogError("Consume error: {Reason}", ex.Error.Reason);
                    throw;
                }
                catch (Exception ex)
                {
                    _logger.LogError("Unexpected error: {Message}", ex.Message);
                }
            }
        }

        private IMessageHandler? GetHandlerForTopic(string topic)
        {
            return topic switch
            {
                "create_post" => _serviceProvider.GetRequiredService<CreatePostHandler>(),
                "delete_post" => _serviceProvider.GetRequiredService<DeletePostHandler>(),
                _ => null,
            };
        }
    }
}

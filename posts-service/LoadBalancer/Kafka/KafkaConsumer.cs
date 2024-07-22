using LoadBalancer.Services;
using Confluent.Kafka;

namespace LoadBalancer.Kafka
{
    public class KafkaConsumer
    {
        private readonly string _bootstrapServers;
        private readonly string _topic;
        private readonly string _groupId;

        private readonly PostServiceClient _postServiceClient;

        public KafkaConsumer(string bootstrapServers, string topic, string groupId, PostServiceClient postServiceClient)
        {
            _bootstrapServers = bootstrapServers;
            _topic = topic;
            _groupId = groupId;
            _postServiceClient = postServiceClient;
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
                    Console.WriteLine("Subscribed to topic.");
                    await ConsumeMessagesAsync(consumer);
                }
                catch (ConsumeException e)
                {
                    Console.WriteLine($"Consume error: {e.Error.Reason}");
                    if (e.Error.Code == ErrorCode.UnknownTopicOrPart)
                    {
                        Console.WriteLine($"Topic '{_topic}' does not exist. Waiting for it to become available...");
                    }
                    await Task.Delay(5000);
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"Unexpected error: {ex.Message}");
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
                    Console.WriteLine($"Consumed message '{consumeResult.Message.Value}' at: '{consumeResult.TopicPartitionOffset}'.");
                    _ = ProcessCreationAsync(consumeResult.Message.Value);
                }
                catch (ConsumeException e)
                {
                    Console.WriteLine($"Consume error: {e.Error.Reason}");
                    throw;
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"Unexpected error: {ex.Message}");
                }
            }
        }

        private async Task ProcessCreationAsync(string message)
        {
            try
            {
                await _postServiceClient.CreatePostAsync(message);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error processing message: {ex.Message}");
            }
        }
    }
}

using LoadBalancer.Services;
using LoadBalancer.Kafka;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Hosting;
using LoadBalancer.Core;

internal class Program
{
    private static async Task Main(string[] args)
    {
        await Task.Delay(10000);

        var host = CreateHostBuilder(args).Build();

        var logger = host.Services.GetRequiredService<ILogger<Program>>();
        logger.LogInformation("Starting Load Balancer");

        var configuration = host.Services.GetRequiredService<IConfiguration>();
        var bootstrapServers = configuration["KAFKA_BROKER"];
        var creationTopic = configuration["CREATE_TOPIC"];
        var deletionTopic = configuration["DELETE_TOPIC"];
        var groupId = configuration["GROUP"];

        var consumerFactory = host.Services.GetRequiredService<KafkaConsumerFactory>();
        var creationTopicConsumer = consumerFactory.CreateConsumer(bootstrapServers, creationTopic, groupId);
        var deletionTopicConsumer = consumerFactory.CreateConsumer(bootstrapServers, deletionTopic, groupId);

        var consumeCreationTask = creationTopicConsumer.ConsumeAsync();  
        var consumeDeletionTask = deletionTopicConsumer.ConsumeAsync();

        await Task.WhenAll(consumeCreationTask,consumeDeletionTask);
    }

    public static IHostBuilder CreateHostBuilder(string[] args) =>
        Host.CreateDefaultBuilder(args)
            .ConfigureAppConfiguration((context, config) =>
            {
                config.AddEnvironmentVariables();
            })
            .ConfigureLogging((context, logging) =>
            {
                logging.ClearProviders();
                logging.AddConsole(options =>
                {
                    options.TimestampFormat = "[yyyy-MM-dd HH:mm:ss] ";
                    options.DisableColors = false;
                });
            })
            .ConfigureServices((context, services) =>
            {
                services.AddSingleton(provider =>
                {
                    var configuration = provider.GetRequiredService<IConfiguration>();
                    var postServiceAdresses = configuration["POST_SERVICE_ADDRESSES"]?.Split(',').ToList();
                    var logger = provider.GetRequiredService<ILogger<PostServiceClient>>();
                    return new PostServiceClient(postServiceAdresses, logger);
                });

                services.AddTransient<KafkaConsumer>();
                services.AddSingleton<KafkaConsumerFactory>();
                services.AddTransient<CreatePostHandler>();
                services.AddTransient<DeletePostHandler>();
            });
}
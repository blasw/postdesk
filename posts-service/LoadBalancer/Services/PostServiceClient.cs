using Grpc.Net.Client;
using Microsoft.Extensions.Logging;

namespace LoadBalancer.Services
{
    public class PostServiceClient
    {
        private readonly List<string> _postServiceAdresses;
        private int _currentServiceIndex;

        private readonly ILogger<PostServiceClient> _logger;


        public PostServiceClient(List<string> addresses, ILogger<PostServiceClient> logger)
        {
            _postServiceAdresses = addresses;
            _currentServiceIndex = 0;
            _logger = logger;
        }

        public async Task CreatePostAsync(string data)
        {
            var address = GetNextServiceAddress();

            using var channel = GrpcChannel.ForAddress(address);
            var client = new PostService.PostServiceClient(channel);

            var response = await client.CreatePostAsync(new CreatePostRequest { JsonData = data });

            _logger.LogInformation("Post creation status: {Status}", (PostCreationStatus)response.Status);
        }

        private string GetNextServiceAddress()
        {
            var address = _postServiceAdresses[_currentServiceIndex];

            _currentServiceIndex = (_currentServiceIndex + 1) % _postServiceAdresses.Count;

            return address;
        }
    }
}

using LoadBalancer.Services;
using Microsoft.Extensions.Logging;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace LoadBalancer.Core
{
    public class CreatePostHandler : IMessageHandler
    {
        private readonly PostServiceClient _postServiceClient;
        private readonly ILogger<CreatePostHandler> _logger;

        public CreatePostHandler(PostServiceClient postServiceClient, ILogger<CreatePostHandler> logger)
        {
            _postServiceClient = postServiceClient;
            _logger = logger;
        }

        public async Task HandleMessageAsync(string message)
        {
            _logger.LogInformation("Handling CreatePost message.");
            await _postServiceClient.CreatePostAsync(message);
        }
    }
}

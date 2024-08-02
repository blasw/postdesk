using LoadBalancer.Services;
using Microsoft.Extensions.Logging;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace LoadBalancer.Core
{
    public class DeletePostHandler : IMessageHandler
    {
        private readonly PostServiceClient _postServiceClient;
        private readonly ILogger<DeletePostHandler> _logger;

        public DeletePostHandler(PostServiceClient postServiceClient, ILogger<DeletePostHandler> logger)
        {
            _postServiceClient = postServiceClient;
            _logger = logger;
        }

        public async Task HandleMessageAsync(string message)
        {
            _logger.LogInformation("Handling DeletePost message.");
            await _postServiceClient.DeletePostAsync(message);
        }
    }
}

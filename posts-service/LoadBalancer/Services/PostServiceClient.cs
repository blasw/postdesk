using Grpc.Net.Client;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace LoadBalancer.Services
{
    public class PostServiceClient
    {
        private readonly List<string> _postServiceAdresses;
        private int _currentServiceIndex;

        public PostServiceClient(List<string> addresses)
        {
            _postServiceAdresses = addresses;
            _currentServiceIndex = 0;
        }

        public async Task CreatePostAsync(string data)
        {
            var address = GetNextServiceAddress();

            using var channel = GrpcChannel.ForAddress(address);
            var client = new PostService.PostServiceClient(channel);

            var response = await client.CreatePostAsync(new CreatePostRequest { JsonData = data });

            Console.WriteLine("Post creation status: " + (PostCreationStatus)response.Status);
        }

        private string GetNextServiceAddress()
        {
            var address = _postServiceAdresses[_currentServiceIndex];

            _currentServiceIndex = (_currentServiceIndex + 1) % _postServiceAdresses.Count;

            return address;
        }
    }
}

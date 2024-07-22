using Grpc.Core;
using Microsoft.Extensions.Hosting;
using Newtonsoft.Json;
using PostService;
using PostService.Models;

namespace PostService.Services
{
    public class PostServiceImpl : PostService.PostServiceBase
    {
        private readonly ILogger<PostServiceImpl> _logger;
        private readonly PostContext _context;
        public PostServiceImpl(ILogger<PostServiceImpl> logger, PostContext context)
        {
            _logger = logger;
            _context = context;
        }

        public override async Task<CreatePostReply> CreatePost(CreatePostRequest request, ServerCallContext context)
        {
            try
            {
                var post = JsonConvert.DeserializeObject<Post>(request.JsonData);

                post.CreatedAt = DateTime.UtcNow;

                _context.Posts.Add(post);
                await _context.SaveChangesAsync();

                _logger.LogInformation("Post created: {Title}", post.Title);

                return new CreatePostReply()
                {
                    Status = (int)PostCreationStatus.Success
                };
            }
            catch (Exception e)
            {
                _logger.LogError("Error creating post : {Message}", e.Message);

                return new CreatePostReply()
                {
                    Status = (int)PostCreationStatus.Error
                };
            }
        }
    }
}

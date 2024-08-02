using Grpc.Core;
using Microsoft.EntityFrameworkCore;
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
                var post = DeserializeJson<Post>(request.JsonData);

                if (post == null)
                {
                    _logger.LogWarning("Invalid post data received.");
                    return CreatePostReply(PostCreationStatus.PostCreationError);
                }

                post.CreatedAt = DateTime.UtcNow;

                _context.Posts.Add(post);
                await _context.SaveChangesAsync();

                _logger.LogInformation("Post created Title: {Title}, Id - {Id}", post.Title, post.Id);

                return CreatePostReply(PostCreationStatus.PostCreationSuccess);
            }
            catch (Exception e)
            {
                _logger.LogError(e, "Error creating post with request data: {RequestData}", request.JsonData);
                return CreatePostReply(PostCreationStatus.PostCreationError);
            }
        }

        public override async Task<DeletePostReply> DeletePost(DeletePostRequest request, ServerCallContext context)
        {
            try
            {
                var deletionInfo = DeserializeJson<DeletePostInfo>(request.JsonData);
                var post = await GetPostByIdAsync(deletionInfo.PostId);

                if (post is null)
                {
                    _logger.LogInformation("Post with ID {PostId} not found.", deletionInfo.PostId);
                    return CreateDeleteReply(PostDeletionStatus.PostDeletionError);
                }

                if (!IsPostAuthor(deletionInfo.UserId, post.AuthorId))
                {
                    _logger.LogWarning("User {User} is not the author of the post {Post}, post author {Author}", deletionInfo.UserId, deletionInfo.PostId, post.AuthorId);
                    return CreateDeleteReply(PostDeletionStatus.PostDeletionError);
                }

                _context.Posts.Remove(post);
                await _context.SaveChangesAsync();

                _logger.LogInformation("Post {PostId} successfully deleted by User {UserId}.", post.Id, deletionInfo.UserId);
                return CreateDeleteReply(PostDeletionStatus.PostDeletionSuccess);
            }
            catch (Exception e)
            {
                _logger.LogError(e, "Error deleting post with request data: {RequestData}", request.JsonData);
                return CreateDeleteReply(PostDeletionStatus.PostDeletionError);
            }
        }

        private T DeserializeJson<T>(string json) where T : new()
        {
            try
            {
                return JsonConvert.DeserializeObject<T>(json) ?? new T();
            }
            catch (JsonException e)
            {
                _logger.LogError(e, "Error deserializing JSON: {Json}", json);
                return new T();
            }
        }

        private Task<Post?> GetPostByIdAsync(int id)
        {
            return _context.Posts.FirstOrDefaultAsync(p => p.Id == id);
        }

        private CreatePostReply CreatePostReply(PostCreationStatus status)
        {
            return new CreatePostReply { Status = (int)status };
        }

        private DeletePostReply CreateDeleteReply(PostDeletionStatus status)
        {
            return new DeletePostReply { Status = (int)status };
        }

        private bool IsPostAuthor(int userId, int authorId)
        {
            return userId == authorId;
        }
    }
}


using Newtonsoft.Json;

namespace PostService.Models
{
    public class Post
    {
        public int Id { get; set; }

        [JsonProperty("title")]
        public string Title { get; set; }

        [JsonProperty("content")]
        public string Content { get; set; }

        [JsonProperty("author_id")]
        public int AuthorId { get; set; }

        public DateTime CreatedAt { get; set; }

        public override string ToString()
        {
            return $"Title: {Title}\n" +
                    $"Content: {Content}\n" +
                    $"AuthorId: {AuthorId} \n" +
                    $"CreatedAt: {CreatedAt}\n";
        }
    }
}

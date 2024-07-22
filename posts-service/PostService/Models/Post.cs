using System.Text.Json.Serialization;

namespace PostService.Models
{
    public class Post
    {
        public int Id { get; set; }

        [JsonPropertyName("title")]
        public string Title { get; set; }

        [JsonPropertyName("content")]
        public string Content { get; set; }

        [JsonPropertyName("author_id")]
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

using Newtonsoft.Json;

namespace PostService.Models
{
    public class DeletePostInfo
    {
        [JsonProperty("post_id")]
        public int PostId { get; set; }

        [JsonProperty("user_id")]
        public int UserId { get; set; }
    }
}

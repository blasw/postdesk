using Microsoft.EntityFrameworkCore;
using PostService.Models;
using PostService.Services;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddGrpc();
builder.Services.AddDbContext<PostContext>(options => options.UseSqlServer(builder.Configuration.GetConnectionString("DefaultConnection"))); 

var app = builder.Build();

// Configure the HTTP request pipeline.
app.MapGrpcService<PostServiceImpl>();
app.MapGet("/", () => "Communication with gRPC endpoints must be made through a gRPC client. To learn how to create a client, visit: https://go.microsoft.com/fwlink/?linkid=2086909");


using (var scope = app.Services.CreateScope())
{
    var dbContext = scope.ServiceProvider.GetRequiredService<PostContext>();
    try
    {
        dbContext.Database.Migrate();
    }
    catch (Exception ex)
    {
        Console.WriteLine($"An error occurred while migrating the database: {ex.Message}");
    }
}


app.Run();

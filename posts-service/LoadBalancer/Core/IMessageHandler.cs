using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace LoadBalancer.Core
{
    public interface IMessageHandler
    {
        Task HandleMessageAsync(string message);
    }

}

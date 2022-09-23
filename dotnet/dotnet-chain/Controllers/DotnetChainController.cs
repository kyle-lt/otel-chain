using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Logging;
using System.Net.Http;

namespace dotnet_chain.Controllers
{
    public class DotnetChainController : ControllerBase
    {
        private static readonly HttpClient client = new HttpClient();
        
        // GET /
        [Route("")]
        [HttpGet]
        public String Home()  {

            return "home";

        }       
        
        // GET /node-chain
        [Route("node-chain")]
        [HttpGet]
        public async Task<String> NodeChain()  {

            // Execute HTTP Client call downstream
            var uri = "http://host.docker.internal:47000/node-chain";
            // Changing to HTTP BIN for now
            //var uri = "http://httpbin.org/get";
            var responseTask = client.GetStringAsync(uri);
            var response = await responseTask;

            return "{\"otel\":\"dotnet\"}";

        }

    }
}

using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Logging;

namespace dotnet_chain.Controllers
{
    public class DotnetChainController : ControllerBase
    {
        // GET /
        [Route("")]
        [HttpGet]
        public String Home()  {

            return "home";

        }       
        
        // GET /node-chain
        [Route("node-chain")]
        [HttpGet]
        public String NodeChain()  {

            return "{\"otel\":\"dotnet\"}";

        }

    }
}

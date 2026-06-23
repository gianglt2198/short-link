# STEPS TO THE MOON 

When I read this requirement, the first thing I always do is to clarify the business workflow. 

# Clarity

- How many characters is the short url? 6 chars 
- What is the maximum of the URL? assume ~2048 chars 
- What is the latency we need to expect for read? assume < 200ms
- Do we need authentication for url? no require for MVP 
- What is the ratio of read to write? expect 50 : 1
- How many urls do we expect to create per day? 100000 requests / day 

There are some scopes I think we need to cover first to design architecture. 

# Back-of-the-envolope and Key Decision

Based on the scopes, I must calculate the volume for both side read and write. 
We need to cover the number of storage and throughput surrounding requests and data.

I assume that 100000 reqs/day (86400) =~ 2 write/s, and x2 at peak ~= 4 write/s.

Then, the ratio of read to write is 50 : 1, 
so the read requests per second is 100 QPS, and approximately 200 QPS. 

After that, I come to the storage of url. the volume of the url at 1KB each is 100MB per day, and we look at about ~36GB. that's fine. 

Based on the analysis, I realize some cases we need to handle. 
- Firstly, the write per second is 2, it can cause the collison if we handle carelessly. Beside, write operation must be strong consistent. 
- Secondly, for read side, to ensure the requirement under 200 ms, eventual consistency is acceptable. the data is staleness. But we need careful handling thundering herd with calling one endpoint so many times, that's called hot spot. 
- Finally, to adapt scalable service for future, we must design ID generator that fit with multiple nodes running in parallel. 

## Key Decision


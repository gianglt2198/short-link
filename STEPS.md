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

### ID Generator

There are some ways to implement ID generator. I will tell a bit pros and cons of these ways:

- Increment ID: it's easiest way to implement ID generator. But, it's too easy to guess and take over from our system.
- Hash URL: it's a bit harder and must use some algorithm to implement, for example, CRC32 (too short), SHA258 (too long) we have to cut only 6 chars from hashed string. it's almost occurred the collision.
- Random 6 chars: generate random 6 chars from base62, it's great to implement that. but sometimes, it causes collision because it's random and we can't control it.
- UUIDv7: actually, I think of UUIDv4 first, I realized that it takes long bit to represent 128bit(16 bytes) and random chars. and I found out that the next generation of UUID is UUIDv7, it includes timestamp and random the rest of bits. If the system is required long ID for distributed system, this is best way to use.
- SnowflakeID: this is greatest way to avoid collision from generating ID simultaniously. And, it only takes 64bit as representing full ID
  - 1 bit: sign (0) - 41 bits: timestamp - 5 bits: data center - 5 bits: node ID - 12 bits: sequence
  - Despite, it can avoid collision in scaling up system, we must handle issues around this solution. it consists of clock skew, coordination.
- Centralized allocation by block: It's like a centralized ID generator distributes a bunch of IDs to each node, and mark them as used. but it's hard to get back ids if the node disrupted or is outage immediately.

So my last decision for ID generator is SnowflakeID. because it avoids collision in parellel process is the best, and we can convert SnowflakeID into base62 simply.
I mitigate the issues:

- the clock skew by saving last_timestamp and compare it with current timestamp; if less a bit ms than, wait until current_time catch up with last_timestamp, and generate; if less some seconds, throw error;
- the coordination by using HPA or auto scaling in K8s, it have node-{number_id}. I can take advange of the number to mark the number of coordinator when generating id.
  In my demo, I only use one node to implement snowflakeID.

### High throughput & Availability

To ensure High throughput & Availability, that is latency under 200ms, we must use cache to reduce request reaching the database. That is the important thing to implement in this system.
Because if the system is under pressure calling URL via API, database might take down and cause the entire system's pending.
We must cover this part as neccessary thing for our system.

# Architecture

![architecture](images/architecture.png)

Architecture is simple.

# Data Model

Table Link: 
- code          : it's for short url (6 chars, unique)
- long_url      : it's for original url 
- created_at    : it's creation date

The most important thing in this model is how to handle collision if it happens.


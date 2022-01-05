# Stats recording and reporting

The Gateway can optionally record the number of hits per function and report it to an external hit counter server.

> This behaviour is disabled by default.

To enable stats reporting, set the `STATS_REPORT_URL` environment variable to the URL of the hit counter server, for example:

    STATS_REPORT_URL=https://www.example.com

Hits are sent periodically as HTTP requests, in the following format:

    PUT /hits/<function name> HTTP/1.1
    Host: www.example.com
    Content-Length: 1
    
    5

The body contains the amount by which to increment the hit counter (for example `5` in the example above).

You can adjust the frequency of stats reporting by setting the `STATS_REPORT_INTERVAL` environment variable to a valid duration, such as `5s` (5 seconds) or `2m` (2 minutes).

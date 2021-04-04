# https-bot

Find HTTP URLs posted on [Hacker News] that can be safely replaced by HTTPS URL.

## FAQ

### Why are you doing this?

Because the "S" in "HTTP" stands for "Secure".

When an user follows an HTTP link,
every hop along the way from their device to the web server can be eavesdropped
and/or manipulated by third parties.

I'm old enough to remember [Firesheep],
how it demonstrated this security risk,
and how it successfully caused Facebook and Twitter to switch to use HTTPS for
everything.

That was 10 years ago.
With the great work of [Let's Encrypt] in the recent years,
we are at a point that for the first time of history,
the majority of websites offer HTTPS support.

This bot is just trying to highlight that,
reminding users that they can follow that link posted on Hacker News securely.

### But I'm using an VPN, why should I care?

Good for you, but an VPN can only help making this connection half secure.
An VPN helps half of the connection, between your device and your VPN endpoint.
The rest half, between your VPN endpoint and web server,
is still not encrypted if you are visiting an HTTP URL.

An HTTPS URL makes it secure all the way, without the need of an VPN.

### This HTTP link actually automatically became HTTPS when I visited it, is this still necessary?

There are a few different reasons that could happen:

1. The domain name is in your browser's preloaded [HSTS] list.
2. You are using some browser extensions, for example [HTTPS Everywhere],
   that rewrote the HTTP request to HTTPS.
3. The web server configured the HTTP to HTTPS redirection.

Although this is certainly much better than web servers not doing
the redirection, no matter whichever the reason it is in your case,
it cannot guarantee that all the requests from all the users are secure:

1. This domain might not be in another browser's preloaded HSTS list.
2. Another user might not be using the same browser extension as you.
3. The first request is still on HTTP that's subject to eavesdropping and
   manipulation.

So it's still better to just post the HTTPS link.

### What does the "with XX% similarity on their contents" part mean?

Although this is unlikely to happen in reality,
in theory someone could configure their web server to serve totally different
contents between HTTP and HTTPS URLs.
In order to avoid misleading users to a URL that's different from what you
intended (the HTTP one you posted),
this bot actually follow both URLs,
read their content (up to 10KiB limit),
and compare the contents read.
It only posts the HTTPS URL if its content is similar enough to the HTTP URL
(the current configured threshold is 95%).

[Hacker News]: https://news.ycombinator.com/
[Firesheep]: https://en.wikipedia.org/wiki/Firesheep
[Let's Encrypt]: https://letsencrypt.org/
[HSTS]: https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security
[HTTPS Everywhere]: https://www.eff.org/https-everywhere

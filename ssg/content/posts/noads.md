---
title: "No Ads for now"
date: 2021-09-07T17:34:04-04:00
draft: true
disable_share: true
---

Well, it would seem I was too early for Ads.

But first, the setup, which was rather straightforward with Hugo despite not being built in.  The first step was creating the `partial`.  That's what Hugo calls a file with an html snippet that gets included into other html files, not to be confused with a `shortcode`, which is also html, but mostly for putting into markdown (I'm probably wrong somewhere, but that's ok for now).  The new `partial` is pretty short, it only has the Google AdSense script snippet they provide:

```html
<script data-ad-client="{{ $.Site.Params.googleAdsense }}" async src="https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js"></script>
```

The stuff between in the `{{ }}`'s tells Hugo to pull the ad client id  from the config file params section like so

```
baseURL = 'https://www.odinseye.cloud'
languageCode = 'en-us'
title = 'Odin'
theme = "ananke"

[params]
  googleAdsense="ca-pub-################"
```

The Ananke theme has a file that gets included into every page called `baseof.html`, and that's where the `partial` gets added.  Here's the file section:

```
...
{{ if eq (getenv "HUGO_ENV") "production" | or (eq .Site.Params.env "production")  }}
    {{ template "_internal/google_analytics.html" . }}
    {{ partial "google_adsense.html" . }}
{{ end }}
...
```

Ananke checks to see if the site is in "production" to add these two bits in, so you don't serve ads to yourself during development.

And that's pretty much all the setup to do from the website side.  The rest is setup in Google AdSense, and waiting to get approved.  Approval (or, in this case, rejection) took about 2 days.  The reason for the rejection was not enough content, which is fair.  TODO: add reasons from google

Anyways, for now I'll disable that script on the site, and move on with adding content.
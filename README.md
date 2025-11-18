# Butterfly

Automate social link preview images, sourced directly from your Web pages.

Butterfly Social is a quick way to auto-generate link preview images (OpenGraph images) in bulk for all your Web pages, without the use of a separate template editor or API integration. The source of truth for the image data & design remains your primary website, so you can use tools you are already familiar with & assets that are already well-integrated into your workflow.

## How to use

1. Create a new hidden element inside your existing Web page, using whatever framework or template engine you use today.
   E.g. here’s a simple example:

    ```html
    <div class="hidden w-[960px]" id="link-preview">
      <h1>Introducing Butterfly Social!</h1>
      <p>Automate social link preview images, sourced directly from your Web pages.</p>
    </div>
    ```

2. Use Butterfly to craft a URL, and paste it into the original page.
    ```html
    <meta property="og:image" content="https://butterfly.your-server.com/link-preview/v1?url=your-site.com/some/page">
    ```

    If the default selector (`#link-preview`) does not work for your use case, provide an alternate one using the `&sel=` parameter.

3. There is no step 3.

Test your Butterfly installation by using your original page URL on any social platform.

Butterfly works well with static sites as well as dynamically-generated sites. If you can put a `<div>` on your page, you can use Butterfly.

### If you can put a `<div>` on your page, you can use Butterfly!

Obviously, that is a bit of a simple example; here’s a more production-worthy example using Tailwind, but you can use (or not use) any Web technologies you want.

```html
<div class="hidden w-[960px]" id="link-preview">
  <div class="flex flex-col h-[540px]">
    <div class="px-16 pt-12 pb-8 text-white background-blue">
      <h1 class="text-shadow-lg text-5xl/18 line-clamp-2">Introducing Butterfly Social</h1>
    </div>
    <p class="my-8 px-16 font-display text-3xl/12 line-clamp-2">
      Automate social link preview images, sourced directly from your Web pages.
    </p>
    <p class="mt-auto px-16 pb-16">
      <img class="inline mr-8 size-16 align-middle" src="/favicon.png">
      <span class="overflow-hidden font-display text-2xl">from Chimbori</span>
    </p>
  </div>
</div>
```

Butterfly captures screenshots at a page scale factor of 2.0 (so you get higher-resolution images, like those on a high-DPI display). Remember to set the width and height of your element to be 0.5 * whatever you want the output to be.

## How it works

1. Butterfly fetches the URL you provide to it, using a Chrome Headless instance;
2. runs JavaScript to un-hide the hidden element;
3. takes a screenshot of it;
4. and serves it
5. (while also caching it).

That’s it.

## How to deploy

We strongly recommend deploying using the official Docker image, which includes Chrome Headless for convenience. Thanks to the [chromedp](https://github.com/chromedp/chromedp) project for making this possible!

- Butterfly is designed to be used behind a TLS reverse proxy for SSL termination (among other things). We recommend [Caddy](https://caddyserver.com/); see sample Caddyfile below.
- If you expect a lot of traffic, consider using a CDN.

### Sample `compose.yml`

  ```yml
  services:
    butterfly:
      container_name: butterfly
      image: ghcr.io/chimbori/butterfly:latest
      volumes:
        - $PWD/butterfly-data:/data
      restart: unless-stopped

  volumes:
    butterfly-data:
  ```

### Sample `butterfly.yml`

To prevent abuse and to conserve resources, Butterfly mandates that you provide an allow-list of domains. Configure this in the required `butterfly.yml` file, and place it in the `/butterfly-data` volume.

```yml
link-preview:
  domains:
    - chimbori.com
    - manas.tungare.name
```

For other configurable options, check out the sample inside the repository.

### Sample `Caddyfile`

```
butterfly.your-server.com {
  reverse_proxy butterfly:9999
  encode zstd gzip
}
```

## Dual Licensed: AGPL & Proprietary

This service is dual-licensed as [Affero GPL](LICENSE.md) (an OSI-approved open-source license) and a Proprietary License.

Feel free (as in freedom!) to install it on your own cheap VPS, as long as you commit to sharing improvements back upstream.

Or, if the AGPL license does not work for your company or organization,
and/or you’d like to support ongoing development and new features,
please [contact us](mailto:hello+butterfly@chimbori.com) for a Proprietary License.

# TODO

Butterfly Social is just getting started; here’s a brief roadmap.

- 🟢 Handle JS errors when selector does not exist; fallback to other strategies
- 🟢 Implement and follow multiple strategies, one by one
  - If selector not found, find title & description to create a screenshot from a default template
  - If no favicon, hide that area
  - If no title, use URL instead
- 🟢 Optimize Chrome resources
- 🟡 Add an Admin UI to enable evicting specific URLs from the cache & other management functions.
- 🟡 Add ACLs & support for multiple projects.
  - Set up PostgreSQL
  - Log access & errors to SQL
- 🔴 Add EXIF info to cover image (to be used as alt text)

(🟢: P0, 🟡: P1, 🔴: P2)

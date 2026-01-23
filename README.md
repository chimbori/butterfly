# Butterfly

Automate social link preview images, sourced directly from your Web pages.

Butterfly Social is a quick way to auto-generate link preview images (OpenGraph images) in bulk for all your Web pages, without the use of a separate template editor or API integration. The source of truth for the image data & design remains your primary website, so you can use tools you are already familiar with & assets that are already well-integrated into your workflow.

## How to use

1. Create a new hidden element inside your existing Web page, using whatever framework or template engine you use today.
   E.g. here’s a [simple example](assets/example.html):

    ```html
    <div id="link-preview" class="hidden w-[960px]">
      <h1>Butterfly Social</h1>
      <p>Automate social link preview images, sourced directly from your Web pages.</p>
    </div>
    ```

2. Use Butterfly to craft a URL, and paste it into the original page.
    ```html
    <meta property="og:image" content="https://butterfly.your-server.com/link-preview/v1?url=your-site.com/some/page">
    ```

    If the default selector (`#link-preview`) does not work for your use case, provide an alternate one using the `&sel=` parameter.

3. There is no step 3.

### How it’s rendered

<img src="assets/example.png" width="800" style="border:1px solid #ccc; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1); border-radius: 24px; margin: 24px 0">

Test your Butterfly installation by posting your original page URL to any social platform.

### If you can put a `<div>` on your page, you can use Butterfly!

## How it works

1. Butterfly fetches the URL you provide to it, using a Chrome Headless instance;
2. runs JavaScript to un-hide the hidden element;
3. takes a screenshot of it;
4. and serves it
5. (while also caching it).

That’s it.

Butterfly works well with static sites (using any static site generator) as well as dynamically-generated sites (using any CMS or platform).

Butterfly captures screenshots at a page scale factor of 2.0 (so you get higher-resolution images, like those on a high-DPI display). Remember to set the width and height of your element to be 0.5 × whatever you want the output to be.

### Can I use…

- Images?

  Yes.

- SVG backgrounds?

  Also, yes.

- Flexbox? Grid?

  Yes, of course.

- Custom fonts? Proprietary fonts?

  Absolutely.

Why limit yourself to the customization possible in a random WYSIWYG editor, when you have the entire Web platform available to you!

Anything you can design for the Web, you can use to create a link preview image.
The infinite is possible at Zombocom. The unattainable is unknown at Zombocom.

## How to deploy

We strongly recommend deploying using the official Docker image, which includes Chrome Headless for convenience. Thanks to the [chromedp](https://github.com/chromedp/chromedp) project for making this possible!

- Butterfly is designed to be used behind a TLS reverse proxy for SSL termination (among other things). We recommend [Caddy](https://caddyserver.com/); see sample Caddyfile below.
- If you expect a lot of traffic, consider using a CDN.

### Sample `compose.yml`

A more advanced example is available in [compose.yml](compose.yml).

```yml
services:
  butterfly:
    container_name: butterfly
    image: ghcr.io/chimbori/butterfly:latest
    volumes:
      - $PWD/butterfly-data:/data
    restart: unless-stopped
    depends_on:
      - butterfly-db

  butterfly-db:
    container_name: butterfly-db
    image: postgres:18-alpine
    environment:
      POSTGRES_DB: butterfly
      POSTGRES_USER: chimbori
      POSTGRES_PASSWORD: chimbori
    volumes:
      - $PWD/butterfly-db-data:/var/lib/postgresql
    restart: unless-stopped

volumes:
  butterfly-data:
  butterfly-db-data:
```

### Sample `butterfly.yml`

Butterfly requires basic configuration to be provided via a config file.
For a full list of configurable options, [see full example](data/butterfly.yml).

- PostgreSQL Database URL
  ```yml
  database:
    url: postgresql://chimbori:chimbori@butterfly-db:5432/butterfly
  ```

- Dashboard credentials (encrypted via `bcrypt`)
  ```yml
  dashboard:
    username: admin
    password: "$2a$10$a8LnUkK1UiB.9yQrUp3wyuGsH1AAHhlHVy1cjIaaIUVAwCtGvaX7q" # "test"
  ```

### Sample `Caddyfile`

```
butterfly.your-server.com {
  reverse_proxy butterfly:9999
  encode zstd gzip
}
```

## Dashboard UI

To prevent abuse and to conserve resources, Butterfly only allows explicitly-allow-listed domains to be used.

You can configure this list using the Dashboard UI at `https://butterfly.your-server.com/dashboard`. The Dashboard is available as an installable PWA (Progressive Web Application) that can be “installed” locally using any modern browser.

<img src="assets/screenshot-pwa.png">

# Dual Licensed: AGPL & Proprietary

This service is dual-licensed as [Affero GPL](LICENSE.md) (an OSI-approved open-source license) and a Proprietary License.

Feel free (as in freedom!) to install it on your own cheap VPS, as long as you commit to sharing improvements back upstream.

Or, if the AGPL license does not work for your company or organization,
and/or you’d like to support ongoing development and new features,
please [contact us](mailto:hello+butterfly@chimbori.com) for a Proprietary License.

# Comparison with Alternatives

There are a lot of paid SaaS tools in this space. Notable among them are:
  - BannerBear
  - RenderForm
  - Templated.io
  - Imejis.io
  - Pablle
  - Orshot
  - Abyssale

They all work roughly the same way: you design a template using their custom tools, then provide them your data (title, description, etc.), and pay them per-request (or per-render) to create & serve those images for you.

This model works great if you do not have access to the source of the page, or have no influence over the developers who build your website.

But now,
- You’ve got to learn a whole new tool.
- That tool exposes a certain amount of design expressiveness, but nowhere near what the Web platform offers natively.
- Anytime you need to change the preview image, you have to visit a completely separate website.
- Anytime your own webpage changes, you have to remember to update the templates to match the theme.
- There’s no way to share themes between your website & these third-party tools: colors, gradients, logos must be copy/pasted manually.
- You have to rely on these companies being around long enough, and not disappearing completely after running out of money or being bought over by a VC.
- And you have to pay, based on volume.

Butterfly is none of those things. All you need is the ability to write some HTML/CSS (no JavaScript necessary!) to design your preview image. And it’s free in perpetuity.

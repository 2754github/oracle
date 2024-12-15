---
tags: ['Next.js', 'React']
---

# React の進化

- [Making Sense of React Server Components](https://www.joshwcomeau.com/react/server-components/)

---

### クライアント側レンダリング（データ取得: クライアント側）

クライアントはサーバーから下記のような HTML を受け取る。

```html
<!DOCTYPE html>
<html>
  <head></head>
  <body>
    <div id="root"></div>
    <script src="bundle.js"></script>
  </body>
</html>
```

bundle.js には**アプリ実行に必要なコード全て**が含まれる。  
bundle.js のロード&実行が完了するまでの間、ユーザーは**空白の画面**を見ることになる。

> [!NOTE]
> このパターンを
>
> - CSR（Client Side Rendering）
>
> と呼ぶことがある。

<br />

リッチな UX を実現できたが、

- バンドルサイズが大きい
- 空白の画面は UX が悪い（当時は SEO 的にも問題があった）

などの課題が残った。

### サーバー側レンダリング + ハイドレーション（データ取得: クライアント側）

クライアントはサーバーから下記のような HTML を受け取る。

```html
<!DOCTYPE html>
<html>
  <head></head>
  <body>
    <!-- ここに、サーバー側で生成された初期HTMLが埋め込まれている。 -->
    <script src="bundle.js"></script>
  </body>
</html>
```

bundle.js には**対話性（※1）やブラウザ API の使用があるコードのみ**が含まれる。  
　※1: 対話性がある = 「useState などのフック」「onClick などのイベントリスナー」の使用がある。  
bundle.js のロード&実行（ハイドレーション）が完了するまでの間、ユーザーは<strong>初期 HTML（※2）</strong>を見ることになる。  
　※2: 例えば、[スケルトンローダー](https://tailwindcss.com/docs/animation#pulse)など。

> [!NOTE]
> 初期 HTML の生成を
>
> - 事前レンダリング（pre-rendering）
>
> と呼ぶことがある。
>
> 事前レンダリングのタイミングが
>
> - ビルド時のものを SSG（Static Site Generation）
> - 再検証時のものを ISR（Incremental Static Regeneration）
> - リクエスト時のものを SSR（Server Side Rendering）
>
> と呼ぶことがある。

> [!WARNING]
> クライアント側でのレンダリングがなくなるわけではないことに注意。例えば、
>
> - state 更新
> - fetch in useEffect で取得したデータでの初期 HTML の置換
>
> などによる再レンダリングは残る。

<br />

「バンドルサイズの削減」と「初期 HTML の用意」はできたが、

- データ取得がクライアント側
- レンダリングがルート単位

などの課題が残った。

### サーバー側レンダリング + ハイドレーション（データ取得: サーバー側）

上記までのパターンと比較すると、データ取得をサーバー側で行った方が効率が良いことが分かる。（下記参照）

```
【クライアント側レンダリング（データ取得: クライアント側）】
Client |    JS DL    | Render |       | Render |
Server |                      | Fetch |        |
                              ▲                ▲
                         First Paint           Content Painted
                          Page Interactive

【サーバー側レンダリング + ハイドレーション（データ取得: クライアント側）】
Client |        | JS DL | Hyd |       | Render |
Server | Render |             | Fetch |        |
                ▲             ▲                ▲
           Frist Paint    Page Interactive     Content Painted

【サーバー側レンダリング + ハイドレーション（データ取得: サーバー側）】
Client |                | JS DL | Hyd |
Server | Fetch | Render |             |
                        ▲             ▲
                   Frist Paint        Page Interactive
                 Content Painted
```

しかし、当時の React ではこのパターンを実現できなかった。

そこで、**各 FW が**独自実装（※1）で、このパターンを実現していた。  
　※1: 例えば、[Next.js の Pages Router の getServerSideProps](https://nextjs.org/docs/pages/building-your-application/data-fetching/get-server-side-props)や[Astro の Islands](https://docs.astro.build/en/concepts/islands/)など。

<br />

「データ取得のサーバー側への移動」はできたが、

- レンダリングがルート単位（持ち越し）
- データ取得がルート単位
- 実装が標準化されていない

などの課題が残った。

### サーバーコンポーネント + ハイドレーション（データ取得: サーバー側）

RSC（[React Server Components](https://ja.react.dev/reference/rsc/server-components)）の登場で、**React で**上記のパターンを実現できるようになった。（実装の標準化）

> [!NOTE]
> コンポーネントは元々サーバー側でもレンダリングされる（例: 事前レンダリング）ため、RSC は RS**o**C（React Server **only** Components）と捉えた方が分かりやすい。

また、RSC はデータ取得を RSC 内で行えるため、データ取得がコンポーネント単位になった。  
例えば、fetch in useEffect を RSC に移動すると、

- コード量の削減
- さらなるバンドルサイズの削減
- クライアント側でのレンダリングの削減（「state 更新による再レンダリング」くらいしかなくなる。）

などが見込める。

さらに、[Streaming HTML and Selective Hydration](https://github.com/reactwg/react-18/discussions/37#:~:text=Streaming%20HTML%20and%20Selective%20Hydration)と組み合わせることで、ルート（**ページ全体**）の事前レンダリングを待つことなく、コンポーネント単位でレンダリング&ストリーミングできるようになった。

> [!NOTE]
> このパターンを
>
> - PPR（[Partial Prerendering](https://nextjs.org/docs/app/building-your-application/rendering/partial-prerendering)）
>
> と呼ぶことがある。

> [!WARNING]
> RSC では対話性やブラウザ API の使用ができないことに注意。  
> これらを使用したい場合、[`'use client'`](https://ja.react.dev/reference/rsc/use-client)で、クライアント/サーバー境界を宣言する必要がある。

<br />

RSC 登場時は「Rails や Laravel に回帰した！」という意見が散見されたが、ルート単位ではなくコンポーネント単位でさまざまな制御ができるなど、全くの別物だと思う。（[この記事](https://speakerdeck.com/mizdra/react-server-components-noyi-wen-wojie-kiming-kasu)が分かりやすい。）

> [!NOTE]
> Rails や Laravel のレンダリング（？）パターンを
>
> - SST（Server Side Templating）
>
> と呼ぶことがある。

# Next.js のレンダリングの挙動

デフォルトで「サーバーコンポーネント + ハイドレーション（データ取得: サーバー側）」になる。

現時点（2024 年 11 月）では experimental だが、[PPR](https://nextjs.org/docs/app/building-your-application/rendering/partial-prerendering)もできる。この場合、`<Suspense>` の `fallback` が事前レンダリングされる。

- [Dynamic Routes](https://nextjs.org/docs/app/building-your-application/routing/dynamic-routes)（`app/blog/[slug]/page.js`）
  - SSR（※1）される。
  - [generateStaticParams](https://nextjs.org/docs/app/api-reference/functions/generate-static-params)を使用することで、SSG（※2）もできる。（※3）
- Static Routes（`app/dashboard/page.js`）（※ 非公式用語）
  - SSG（※2）される。（※3）

※1: Next.js は[Dynamic Rendering](https://nextjs.org/docs/app/building-your-application/rendering/server-components#dynamic-rendering)と呼んでいる。  
※2: Next.js は[Static Rendering (Default)](https://nextjs.org/docs/app/building-your-application/rendering/server-components#static-rendering-default)と呼んでいる。  
※3: レンダリング中に「[Dynamic APIs](https://nextjs.org/docs/app/building-your-application/rendering/server-components#dynamic-apis)」または「キャッシュされていないデータ」が検出されると、ルート（ページ全体）が SSR に切り替わる。（[Switching to Dynamic Rendering](https://nextjs.org/docs/app/building-your-application/rendering/server-components#switching-to-dynamic-rendering)）

> [!TIP]
> データをキャッシュするには？
>
> `'use cache'`実装中で情報が錯綜している。
>
> - [dynamicIO](https://nextjs.org/docs/app/api-reference/next-config-js/dynamicIO)
> - [use cache](https://nextjs.org/docs/app/api-reference/directives/use-cache)
>
> 現時点（2024 年 11 月）での参考
>
> - [Caching in Next.js](https://nextjs.org/docs/app/building-your-application/caching)
> - [Data Fetching and Caching](https://nextjs.org/docs/app/building-your-application/data-fetching/fetching)
> - [Incremental Static Regeneration (ISR)](https://nextjs.org/docs/canary/app/building-your-application/data-fetching/incremental-static-regeneration)

# Next.js のコンポーネントの挙動

- [https://x.com/d151005/status/1844869688796512561](https://x.com/d151005/status/1844869688796512561)

---

デフォルトで RSC になる。（[Using Server Components in Next.js](https://nextjs.org/docs/app/building-your-application/rendering/server-components#using-server-components-in-nextjs)）

> [!NOTE]
> 実際にはサーバーでもクライアントでも使用できるコンポーネント（Shared Components）になる。

RSC では下記（RSC Payload）がクライアントに返る。（[How are Server Components rendered?](https://nextjs.org/docs/app/building-your-application/rendering/server-components#how-are-server-components-rendered)）  
（1）RSC のレンダリング結果（HTML）  
（2）BC のスロットとそのバンドル（JS）への参照（※ RSC が BC を含む場合）  
（3）BC に渡される props 全て（シリアライズされたデータ）（※ RSC が BC を含む場合）

> [!WARNING]
>
> `'use client'`が宣言されたコンポーネントを BC（Boundary Components）と表記する。（※非公式用語）  
> Next.js は Client Components と呼んでいるが、誤解を生むため、好まない。
>
> あくまでも`'use client'`は「モジュールのクライアント/サーバー境界」を宣言するもので、「コンポーネントがどこでレンダリングされるか」とは無関係。  
> 「モジュールの依存関係」と「コンポーネントの親子関係」を混同しないように注意。

BC では下記がクライアントに返る。（[How are Client Components Rendered?](https://nextjs.org/docs/app/building-your-application/rendering/client-components#how-are-client-components-rendered)）  
（a）BC のレンダリング結果（HTML）  
（b）ハイドレーション用のバンドル（JS）  
（c）import したもの全て（JS）

# Next.js のコンポーネントの書き方

### RSC

- （2）より、**RSC は BC を import して良い**。
  - ただし、（3）より、BC に渡す props は[シリアライズ可能](https://ja.react.dev/reference/rsc/use-server#serializable-parameters-and-return-values)である必要がある。また、**BC に渡す props にクレデンシャルを含めてはいけない**。

### BC

- （b）より、**BC はクレデンシャルにアクセスする関数を import してはいけない**。
  - ただし、[Server Actions](https://ja.react.dev/reference/rsc/use-server)であれば import して良い。
    - Server Actions はクライアントに返るのではなく、HTTP エンドポイント化されるため。
    - Server Actions を使用することで、BC を減らすことができる。
      ```jsx
      // 'use client' 必要
      <button onClick={login}>ログイン</button>
      ```
      ```jsx
      // 'use client' 不要
      <form>
        <button formAction={login}>ログイン</button>
      </form>
      ```
- （c）より、BC は RSC を import してしまうと、せっかくの RSC がクライアント側でレンダリングされてしまうため、**BC は RSC を import するのではなく、props として受け取るべき**。（[Supported Pattern: Passing Server Components to Client Components as Props](https://nextjs.org/docs/app/building-your-application/rendering/composition-patterns#supported-pattern-passing-server-components-to-client-components-as-props)）
  - そもそも[「BC は RSC を import できない」との記載](https://nextjs.org/docs/app/building-your-application/rendering/composition-patterns#unsupported-pattern-importing-server-components-into-client-components)があるが、import しても正常に動作する。おそらく正しくは「BC は RS**o**C を import できない」である。
    - サーバー専用 API を使用すると、自動的に RS**o**C になる。
    - `import 'server-only'`を使用すると、強制的に RS**o**C にできる。（[Keeping Server-only Code out of the Client Environment](https://nextjs.org/docs/app/building-your-application/rendering/composition-patterns#:~:text=npm%20install%20server%2Donly)）
    - **RSoC に限らず、クレデンシャルにアクセスする関数や Server Actions などでも、`import 'server-only'`を使用すべき**。

### 全体

- **著しく可読性を下げる最適化（過度な RSC と BC の分離）は避けるべき**。

# 積読

- Suspense + use: [React 19: using server promises in client components](https://x.com/alfonsusac/status/1787349893205639646)
- [PPR - pre-rendering 新時代の到来と SSR/SSG 論争の終焉](https://zenn.dev/akfm/articles/nextjs-partial-pre-rendering)
- [PPR はアイランドアーキテクチャなのか](https://zenn.dev/akfm/articles/ppr-vs-islands-architecture)
- [Next.js の "use cache" ディレクティブによるキャッシュ制御](https://azukiazusa.dev/blog/cache-control-with-use-cache-directive-in-nextjs/)
- [なぜ Server Actions を使うのか](https://azukiazusa.dev/blog/why-use-server-actions/)
- [知らないとあぶない、Next.js セキュリティばなし](https://zenn.dev/moozaru/articles/d270bbc476758e)
- ["use server"; で export した関数が意図せず？公開される](https://zenn.dev/moozaru/articles/b0ef001e20baaf)
- ["use server";を勘違いして使うと危ない](https://zenn.dev/moozaru/articles/c7335f66dfb8df)
- [Next.js で簡単な CRUD アプリを作りながら気になったセキュリティ: Rails の視点から](https://zenn.dev/naofumik/articles/c699deb688ac04)

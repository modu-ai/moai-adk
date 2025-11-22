# Frontend Performance Optimization

## Bundle Size Optimization

### Code Splitting Strategies

```typescript
// next.config.js
module.exports = {
    webpack: (config, { isServer }) => {
        config.optimization.splitChunks.cacheGroups = {
            default: false,
            vendors: false,
            // Vendor code
            vendor: {
                filename: 'chunks/vendor.js',
                chunks: 'all',
                test: /node_modules/,
                priority: 20
            },
            // React
            common: {
                minChunks: 2,
                priority: 10,
                reuseExistingChunk: true,
                name: 'common'
            }
        };

        return config;
    }
};
```

### Dynamic Imports

```typescript
// components/HeavyComponent.tsx
import dynamic from 'next/dynamic';

const HeavyComponent = dynamic(() => import('./Heavy'), {
    loading: () => <div>Loading...</div>,
    ssr: false
});

export default function Page() {
    return <HeavyComponent />;
}
```

## Image Optimization

### Next.js Image Component

```typescript
import Image from 'next/image';
import img from '@/public/hero.webp';

export function HeroImage() {
    return (
        <Image
            src={img}
            alt="Hero"
            priority
            placeholder="blur"
            sizes="(max-width: 768px) 100vw, 50vw"
        />
    );
}
```

## CSS Performance

### CSS-in-JS Optimization (Styled Components)

```typescript
import styled from 'styled-components';

const Button = styled.button`
    background: ${props => props.primary ? 'blue' : 'gray'};
    padding: 10px 20px;
    border: none;
    cursor: pointer;

    &:hover {
        opacity: 0.8;
    }

    @media (max-width: 768px) {
        padding: 8px 16px;
    }
`;

// Babel plugin for smaller bundle
// .babelrc: { "plugins": ["babel-plugin-styled-components"] }
```

## Web Vitals Optimization

### Core Web Vitals Monitoring

```typescript
// lib/vitals.ts
import { onCLS, onFCP, onFID, onINP, onLCP } from 'web-vitals/attribution';

export function reportWebVitals() {
    onCLS(console.log);
    onFCP(console.log);
    onFID(console.log);
    onINP(console.log);
    onLCP(console.log);
}

// pages/_app.tsx
import { reportWebVitals } from '@/lib/vitals';

export default function App({ Component, pageProps }) {
    useEffect(() => {
        reportWebVitals();
    }, []);

    return <Component {...pageProps} />;
}
```

## JavaScript Performance

### useCallback and useMemo Optimization

```typescript
import { useCallback, useMemo } from 'react';

interface DataGridProps {
    data: Item[];
}

export function DataGrid({ data }: DataGridProps) {
    // Memoize expensive computation
    const sortedData = useMemo(
        () => [...data].sort((a, b) => a.name.localeCompare(b.name)),
        [data]
    );

    // Memoize callback to prevent child re-renders
    const handleRowClick = useCallback((rowId: string) => {
        console.log('Row clicked:', rowId);
    }, []);

    return (
        <table>
            <tbody>
                {sortedData.map(item => (
                    <tr key={item.id} onClick={() => handleRowClick(item.id)}>
                        <td>{item.name}</td>
                    </tr>
                ))}
            </tbody>
        </table>
    );
}
```

## Server-Side Rendering Optimization

### ISR (Incremental Static Regeneration)

```typescript
// pages/blog/[slug].tsx
import { GetStaticProps, GetStaticPaths } from 'next';

export async function getStaticPaths(): Promise<GetStaticPaths> {
    const posts = await getPosts();
    return {
        paths: posts.map(post => ({
            params: { slug: post.slug }
        })),
        fallback: 'blocking'
    };
}

export async function getStaticProps({ params }: GetStaticPropsContext) {
    const post = await getPostBySlug(params.slug);

    return {
        props: { post },
        revalidate: 3600 // Revalidate every hour
    };
}
```

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready

## Context7 Integration

### Performance Tools & Libraries
- [Next.js Image Optimization](/vercel/next.js): Native image optimization
- [web-vitals](/GoogleChrome/web-vitals): Performance metrics
- [Lighthouse](/GoogleChrome/lighthouse): Performance auditing

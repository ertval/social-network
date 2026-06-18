'use client';

/**
 * app/page.tsx  →  route: /
 *
 * Mirrors renderHomePage() from pages/home.js.
 * Protected — redirects to /login if not authenticated.
 */

import { useEffect, useState } from 'react';
import Link from 'next/link';
import AuthGuard from '@/components/AuthGuard';
import { CategoryList, CategoriesSkeleton } from '@/components/CategoryCard';
import { fetchCategories } from '@/lib/api';
import { prepareCategories } from '@/lib/helpers';
import type { Category } from '@/lib/types';

export default function HomePage() {
  return (
    <AuthGuard>
      <HomeContent />
    </AuthGuard>
  );
}

function HomeContent() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  function normalizeCategoriesData(data: unknown): Record<string, unknown>[] {
    if (Array.isArray(data)) {
      return data as Record<string, unknown>[];
    }

    if (data && typeof data === 'object') {
      const obj = data as Record<string, unknown>;

      // Try common property names
      const categoriesData = obj.categories ?? obj.Categories ?? obj.data ?? obj.Items;

      if (Array.isArray(categoriesData)) {
        return categoriesData as Record<string, unknown>[];
      }
    }

    return [];
  }

  useEffect(() => {
    async function load() {
      try {
        const data = await fetchCategories();
        const rawArray = normalizeCategoriesData(data);
        setCategories(prepareCategories(rawArray));
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load categories');
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  // Close the <details> dropdown on outside click — mirrors initCategoryDetails()
  useEffect(() => {
    const details = document.querySelector('.category-details');
    if (!details) return;
    function handleClick(e: MouseEvent) {
      if (!details!.contains(e.target as Node)) details!.removeAttribute('open');
    }
    document.addEventListener('click', handleClick, { passive: true });
    return () => document.removeEventListener('click', handleClick);
  }, [categories]);

  if (loading) return <HomeSkeleton />;
  if (error) return <HomeError message={error} />;

  return (
    <>
      <h1 className="forum-title">Welcome to Forum</h1>
      <div className="main-container">
        {/* Category nav + dropdown — mirrors buildCategoryDetailsHTML() */}
        <div className="nav-categories">
          <details className="category-details">
            <summary>Categories</summary>
            <div className="details-content">
              {categories.length > 0 ? (
                categories.map((cat) => {
                  const id = String(cat.ID ?? cat.id ?? '');
                  const name = cat.Name ?? cat.name ?? '';
                  const color = cat.Color ?? cat.color ?? '#00C6FF';
                  const topicCount = cat.TopicCount ?? cat.topic_count ?? 0;
                  return (
                    <Link
                      key={id}
                      href={`/topics?search=&category=${id}`}
                      className="details-category-link"
                    >
                      <div className="details-text-box">
                        <span className="category-title-color" style={{ backgroundColor: color }} />
                        <span className="details-category-title">{name}</span>
                      </div>
                      <span className="category-count">{topicCount}</span>
                    </Link>
                  );
                })
              ) : (
                <span className="details-category-title">No categories yet</span>
              )}
            </div>
          </details>

          <Link href="/categories" className="nav-categories-btn">
            Categories
          </Link>
          <Link href="/topics" className="nav-categories-btn">
            Topics
          </Link>
        </div>

        {/* Shared category cards */}
        <CategoryList
          categories={categories}
          emptyMessage="No categories found. Check back later!"
        />
      </div>
    </>
  );
}

// ─── Skeleton ─────────────────────────────────────────────────────────────────

function HomeSkeleton() {
  return (
    <>
      <h1 className="forum-title">Welcome to Forum</h1>
      <div className="main-container">
        <div className="nav-categories">
          <div
            className="skeleton skeleton-btn"
            style={{ width: 110, height: 40, borderRadius: 4 }}
          />
          <div
            className="skeleton skeleton-btn"
            style={{ width: 100, height: 40, borderRadius: 4 }}
          />
          <div
            className="skeleton skeleton-btn"
            style={{ width: 80, height: 40, borderRadius: 4 }}
          />
        </div>
        <CategoriesSkeleton count={3} />
      </div>
    </>
  );
}

// ─── Error state ──────────────────────────────────────────────────────────────

function HomeError({ message }: { message: string }) {
  return (
    <>
      <h1 className="forum-title">Welcome to Forum</h1>
      <div className="main-container">
        <div className="categories-container" style={{ padding: '2rem', textAlign: 'center' }}>
          <p style={{ color: '#e53e3e', fontSize: '1.1rem' }}>⚠️ {message}</p>
          <p style={{ marginTop: '1rem', color: 'var(--grey-color)' }}>
            Could not load categories. Please try refreshing the page.
          </p>
        </div>
      </div>
    </>
  );
}

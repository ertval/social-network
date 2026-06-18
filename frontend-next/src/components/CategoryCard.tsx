/**
 * components/CategoryCard.tsx
 *
 * Shared between the home page and the categories page.
 * Mirrors buildCategoriesListHTML() + the CategoryCard render from pages/home.js
 * and the more complete version in the current app/page.tsx.
 */

import Link from 'next/link';
import { formatRelativeDate } from '@/lib/helpers';
import type { Category, Topic } from '@/lib/types';

// ─── CategoryList ─────────────────────────────────────────────────────────────

interface CategoryListProps {
  categories: Category[];
  emptyMessage?: string;
}

export function CategoryList({
  categories,
  emptyMessage = 'No categories found. Check back later!',
}: CategoryListProps) {
  if (!categories.length) {
    return (
      <div className="categories-container">
        <p className="no-topics-message">{emptyMessage}</p>
      </div>
    );
  }

  return (
    <div className="categories-container">
      {categories.map((cat) => (
        <CategoryCard key={cat.ID ?? cat.id} category={cat} />
      ))}
    </div>
  );
}

// ─── CategoryCard ─────────────────────────────────────────────────────────────

export function CategoryCard({ category: cat }: { category: Category }) {
  const id = String(cat.ID ?? cat.id ?? '');
  const name = cat.Name ?? cat.name ?? '';
  const description = cat.Description ?? cat.description ?? '';
  const color = cat.Color ?? cat.color ?? '#00C6FF';
  const imagePath = cat.ImagePath ?? cat.image_path ?? '/images/categories/default_category.png';

  const topics: Topic[] = Array.isArray(cat.Topics)
    ? cat.Topics
    : Array.isArray(cat.topics)
      ? cat.topics
      : [];

  return (
    <div className="category">
      <div className="category-wrapper">
        <div className="category-img-box">
          <Link href={`/topics?search=&category=${id}`} className="category-link">
            {/* Regular <img> — images come from the Go backend, not Next.js static */}
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="category-img" src={imagePath} alt={name} />
          </Link>
        </div>
        <div className="category-info">
          <div className="category-info-box">
            <Link href={`/topics?search=&category=${id}`} className="category-link">
              <div className="category-title-box">
                <span className="category-title-color" style={{ backgroundColor: color }} />
                <span className="category-title">{name}</span>
              </div>
            </Link>
            <p className="category-description">{description}</p>
          </div>
        </div>
      </div>

      {/* Latest topic previews inside the card */}
      <div className="category-posts">
        {topics.length > 0 ? (
          topics.slice(0, 3).map((topic) => {
            const topicID = String(topic.ID ?? topic.id ?? '');
            const topicTitle = topic.Title ?? topic.title ?? 'Untitled';
            const topicDate = formatRelativeDate(topic.CreatedAt ?? topic.created_at ?? '');

            return (
              <div key={topicID} className="category-post">
                <Link href={`/topic/${topicID}`} className="topic-link">
                  <span className="category-post-title">
                    <span className="left-arrow">&#10147;</span> {topicTitle}
                  </span>
                </Link>
                <span className="category-post-date">{topicDate}</span>
              </div>
            );
          })
        ) : (
          <span className="category-post-date">No posts yet</span>
        )}
      </div>
    </div>
  );
}

// ─── Skeleton (used by both home and categories) ──────────────────────────────

export function CategoriesSkeleton({ count = 3 }: { count?: number }) {
  return (
    <div className="categories-container">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="category category--skeleton">
          <div className="category-wrapper">
            <div className="skeleton skeleton-img" />
            <div className="category-info-box">
              <div className="skeleton skeleton-title" />
              <div className="skeleton skeleton-desc" />
            </div>
          </div>
          <div className="category-posts">
            <div className="skeleton skeleton-post" />
            <div className="skeleton skeleton-post" />
            <div className="skeleton skeleton-post" />
          </div>
        </div>
      ))}
    </div>
  );
}

import { useDeferredValue, useEffect, useState } from 'react';
import { CalendarDays, Code2, MessageSquareText, Search, Star, Trophy, UsersRound } from 'lucide-react';
import ReactMarkdown from 'react-markdown';

import { auditPublicHub, fetchPublicCommunity, type CommunityPost, type PublicCommunity } from '../api';

const filters = [
  { id: 'all', label: 'All content' },
  { id: 'discussion', label: 'Discussions' },
  { id: 'tutorial', label: 'Tutorials & docs' },
  { id: 'project', label: 'Projects' },
  { id: 'announcement', label: 'Announcements' },
] as const;

type CommunityPageProps = {
  onLogin: () => void;
  onSignup: () => void;
};

function CommunityMedia({ post }: { post: CommunityPost }) {
  if (!post.media_url) return null;
  try {
    const url = new URL(post.media_url);
    if (url.protocol !== 'https:') return null;
    if (/\.(png|jpe?g|webp|gif)$/i.test(url.pathname)) {
      return <img alt="" className="community-post__media" loading="lazy" src={url.toString()} />;
    }
    const youtubeId = url.hostname.includes('youtube.com') ? url.searchParams.get('v') : url.hostname === 'youtu.be' ? url.pathname.slice(1) : '';
    if (youtubeId && /^[a-zA-Z0-9_-]{6,20}$/.test(youtubeId)) {
      return <iframe allow="accelerometer; encrypted-media; picture-in-picture" className="community-post__media" loading="lazy" sandbox="allow-scripts allow-same-origin allow-presentation" src={`https://www.youtube-nocookie.com/embed/${youtubeId}`} title={`${post.title} media`} />;
    }
  } catch {
    return null;
  }
  return null;
}

export function CommunityPage({ onLogin, onSignup }: CommunityPageProps) {
  const [community, setCommunity] = useState<PublicCommunity | null>(null);
  const [activeFilter, setActiveFilter] = useState<(typeof filters)[number]['id']>('all');
  const [query, setQuery] = useState('');
  const [expandedPost, setExpandedPost] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const deferredQuery = useDeferredValue(query.trim().toLowerCase());

  useEffect(() => {
    const controller = new AbortController();
    void fetchPublicCommunity(controller.signal)
      .then((response) => { setCommunity(response); setError(null); })
      .catch((loadError) => { if (!controller.signal.aborted) setError(loadError instanceof Error ? loadError.message : 'Community data unavailable'); });
    return () => controller.abort();
  }, []);

  const posts = (community?.posts ?? []).filter((post) => {
    const filterMatches = activeFilter === 'all' || post.kind === activeFilter;
    const queryMatches = deferredQuery === '' || `${post.title} ${post.excerpt} ${post.markdown} ${post.author_username}`.toLowerCase().includes(deferredQuery);
    return filterMatches && queryMatches;
  });
  const featured = (community?.posts ?? []).filter((post) => post.featured).slice(0, 3);

  function togglePost(post: CommunityPost) {
    const willOpen = expandedPost !== post.id;
    setExpandedPost(willOpen ? post.id : null);
    if (willOpen) void auditPublicHub('community-post', post.id);
  }

  return (
    <>
      <section className="community-hero">
        <div className="community-hero__copy">
          <span className="eyebrow">OrcaStack community</span>
          <h1>Build the control plane with the people who operate it.</h1>
          <p>Discussions, implementation guides, project work, and platform announcements in one searchable developer hub.</p>
          <div className="community-hero__actions"><button className="primary-button primary-button--warm" onClick={onSignup} type="button"><UsersRound aria-hidden="true" size={17} /> Join community</button><button className="secondary-button" onClick={onLogin} type="button">Contributor sign in</button></div>
        </div>
        <div className="community-search">
          <Search aria-hidden="true" size={19} />
          <label htmlFor="community-search">Search community content</label>
          <input id="community-search" onChange={(event) => setQuery(event.target.value)} placeholder="Search tutorials, discussions, projects..." type="search" value={query} />
          <span>{community?.posts.length ?? 0} published posts</span>
        </div>
      </section>

      {error ? <div className="public-data-notice">{error}</div> : null}

      <section className="community-featured">
        <div className="landing-section__heading"><div><span className="eyebrow">Featured now</span><h2>Projects and knowledge selected by maintainers.</h2></div></div>
        <div className="community-featured__grid">
          {featured.length === 0 ? <p className="community-empty">No featured community content has been published.</p> : featured.map((post) => (
            <button key={post.id} onClick={() => togglePost(post)} type="button"><span><Star aria-hidden="true" size={15} /> {post.kind}</span><strong>{post.title}</strong><small>{post.excerpt}</small></button>
          ))}
        </div>
      </section>

      <section className="community-layout">
        <div className="community-feed">
          <div className="community-filters" role="group" aria-label="Filter community content">
            {filters.map((filter) => <button aria-pressed={activeFilter === filter.id} className={activeFilter === filter.id ? 'community-filter community-filter--active' : 'community-filter'} key={filter.id} onClick={() => setActiveFilter(filter.id)} type="button">{filter.label}</button>)}
          </div>
          <div className="community-posts">
            {posts.length === 0 ? <p className="community-empty">No community content matches this search.</p> : posts.map((post) => (
              <article className={expandedPost === post.id ? 'community-post community-post--expanded' : 'community-post'} key={post.id}>
                <button className="community-post__summary" onClick={() => togglePost(post)} type="button">
                  <span className="community-post__kind">{post.kind === 'tutorial' ? <Code2 aria-hidden="true" size={15} /> : <MessageSquareText aria-hidden="true" size={15} />}{post.kind}</span>
                  <h3>{post.title}</h3><p>{post.excerpt}</p><small>By {post.author_username} · {new Date(post.updated_at).toLocaleDateString()}</small>
                </button>
                {expandedPost === post.id ? <div className="community-post__body"><CommunityMedia post={post} /><ReactMarkdown>{post.markdown}</ReactMarkdown></div> : null}
              </article>
            ))}
          </div>
        </div>

        <aside className="community-sidebar">
          <section><span className="eyebrow"><Trophy aria-hidden="true" size={14} /> Contributor leaderboard</span><ol className="contributor-list">{(community?.contributors ?? []).map((contributor, index) => <li key={contributor.username}><span>{index + 1}</span><div><strong>{contributor.display_name}</strong><small>@{contributor.username}</small></div><b>{contributor.contributions}</b></li>)}</ol></section>
          <section><span className="eyebrow"><CalendarDays aria-hidden="true" size={14} /> Upcoming events</span>{(community?.events ?? []).length === 0 ? <p className="community-empty">No upcoming events are published.</p> : <ul className="community-events">{community?.events.map((event) => <li key={event.id}><time>{new Date(event.starts_at).toLocaleDateString()}</time><strong>{event.title}</strong><span>{event.location}</span></li>)}</ul>}</section>
        </aside>
      </section>
    </>
  );
}

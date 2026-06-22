import { useEffect, useState, useMemo } from "react";
import ReactMarkdown from "react-markdown";
import { useI18nContext } from "../i18n/I18nProvider";

interface InfoPageProps {
  page: "about" | "privacy";
}

async function fetchMarkdown(page: string, locale: string): Promise<string> {
  const resp = await fetch(`/info-pages/${locale}/${page}.md`);
  if (!resp.ok) {
    throw new Error(`Failed to load markdown: ${resp.status}`);
  }
  return resp.text();
}

function extractFirstH1(markdown: string): string {
  for (const line of markdown.split("\n")) {
    const trimmed = line.trim();
    if (trimmed.startsWith("# ")) {
      return trimmed.slice(2).trim();
    }
  }
  return "";
}

export default function InfoPage({ page }: InfoPageProps) {
  const { locale } = useI18nContext();
  const [content, setContent] = useState<string>("");
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const text = await fetchMarkdown(page, locale);
        if (!cancelled) {
          setContent(text);
          setNotFound(false);
        }
      } catch {
        try {
          const text = await fetchMarkdown(page, "en");
          if (!cancelled) {
            setContent(text);
            setNotFound(false);
          }
        } catch {
          if (!cancelled) {
            setContent("");
            setNotFound(true);
          }
        }
      }
    }

    load();

    return () => {
      cancelled = true;
    };
  }, [page, locale]);

  const title = useMemo(
    () => (content ? extractFirstH1(content) || page : page),
    [content, page]
  );

  if (notFound) {
    return (
      <main style={{ maxWidth: "48rem", margin: "0 auto", padding: "3rem 1.5rem" }}>
        <h1 style={{ fontSize: "1.5rem", fontWeight: 700, marginBottom: "1rem", color: "var(--color-base-font)" }}>
          Page not found
        </h1>
      </main>
    );
  }

  return (
    <main style={{ maxWidth: "48rem", margin: "0 auto", padding: "3rem 1.5rem" }}>
      <h1 style={{ fontSize: "1.5rem", fontWeight: 700, marginBottom: "1rem", color: "var(--color-base-font)" }}>
        {title}
      </h1>
      <div style={{ color: "var(--color-base-font)", lineHeight: 1.6 }}>
        <ReactMarkdown
          components={{
            a: (props) => <a {...props} style={{ color: "var(--color-result-link)" }} />,
          }}
        >
          {content}
        </ReactMarkdown>
      </div>
    </main>
  );
}

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";

const categoryKeys = ["general", "images", "videos", "news", "map", "music", "it", "science", "files", "social_media"];

export default function Sidebar() {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  return (
    <>
      {/* Hamburger */}
      <button
        onClick={() => setOpen(!open)}
        className="fixed top-4 start-4 z-50 p-2 rounded-lg bg-[#1a1a1a] text-[#9ca3af] hover:text-white hover:bg-[#2a2a2a] transition-colors"
        aria-label="Toggle sidebar"
      >
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>

      {/* Overlay */}
      {open && (
        <div className="fixed inset-0 z-40 bg-black/50" onClick={() => setOpen(false)} />
      )}

      {/* Sidebar */}
      <aside
        className={`fixed top-0 start-0 z-50 h-full w-64 bg-[#1a1a1a] border-e border-[rgba(255,255,255,0.08)] 
                    transform transition-transform duration-300 ease-in-out
                    ${open ? "translate-x-0" : "-translate-x-full"}`}
      >
        <div className="p-6">
          <h2 className="text-lg font-semibold text-white mb-6">
            <span className="text-[#3b82f6]">Sear</span>Go
          </h2>

          {/* Categories */}
          <div className="mb-6">
            <h3 className="text-xs uppercase tracking-wider text-[#6b7280] mb-3">{t("sidebar.categories")}</h3>
            <div className="space-y-1">
              {categoryKeys.map((cat) => (
                <Link
                  key={cat}
                  to={`/?category=${cat}`}
                  onClick={() => setOpen(false)}
                  className="block px-3 py-2 rounded-lg text-sm text-[#d1d5db] hover:bg-[#2a2a2a] hover:text-white transition-colors"
                >
                  {t(`category.${cat}`)}
                </Link>
              ))}
            </div>
          </div>

          {/* Settings */}
          <div className="mb-6">
            <h3 className="text-xs uppercase tracking-wider text-[#6b7280] mb-3">{t("sidebar.settings")}</h3>
            <Link
              to="/settings"
              onClick={() => setOpen(false)}
              className="block px-3 py-2 rounded-lg text-sm text-[#d1d5db] hover:bg-[#2a2a2a] hover:text-white transition-colors"
            >
              {t("sidebar.settings")}
            </Link>
          </div>

          {/* Links */}
          <div>
            <h3 className="text-xs uppercase tracking-wider text-[#6b7280] mb-3">Links</h3>
            <Link
              to="/about"
              onClick={() => setOpen(false)}
              className="block px-3 py-2 rounded-lg text-sm text-[#d1d5db] hover:bg-[#2a2a2a] hover:text-white transition-colors"
            >
              {t("sidebar.about")}
            </Link>
          </div>
        </div>
      </aside>
    </>
  );
}

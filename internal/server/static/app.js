// URL state helpers
function readURLState() {
  const params = new URLSearchParams(window.location.search);
  const hash = window.location.hash;
  return {
    paper: hash.startsWith("#paper=") ? hash.slice(7) : null,
    category: params.get("cat") || null,
    sort: params.get("sort") || "date-desc",
    q: params.get("q") || "",
  };
}

function writeURLState() {
  const params = new URLSearchParams();
  if (selectedCategory) params.set("cat", selectedCategory);
  if (sortOrder !== "date-desc") params.set("sort", sortOrder);
  if ($search.value) params.set("q", $search.value);
  const qs = params.toString();
  const hash = selectedPaperID ? `#paper=${selectedPaperID}` : "";
  const url = qs ? `?${qs}${hash}` : `${window.location.pathname}${hash}`;
  history.replaceState(null, "", url);
}

// State
let papers = [];
let categories = [];
let selectedCategory = null;
let selectedPaperID = null;
let showJapanese = false;
let sortOrder = "date-desc";

// Mobile helpers
function isMobile() {
  return window.matchMedia("(max-width: 640px)").matches;
}

// DOM
const $app = document.getElementById("app");
const $sidebar = document.getElementById("sidebar");
const $paperList = document.getElementById("paper-list");
const $categories = document.getElementById("categories");
const $papers = document.getElementById("papers");
const $search = document.getElementById("search");
const $paperCount = document.getElementById("paper-count");
const $sortSelect = document.getElementById("sort-select");
const $detailEmpty = document.getElementById("detail-empty");
const $detailContent = document.getElementById("detail-content");
const $langToggle = document.getElementById("lang-toggle");
const $btnEn = document.getElementById("btn-en");
const $btnJa = document.getElementById("btn-ja");
const $toolbarId = document.getElementById("toolbar-id");
const $btnArxiv = document.getElementById("btn-arxiv");
const $btnDownload = document.getElementById("btn-download");
const $btnToggleMeta = document.getElementById("btn-toggle-meta");
const $btnToggleSidebar = document.getElementById("btn-toggle-sidebar");
const $btnToggleList = document.getElementById("btn-toggle-list");
const $pdfFrame = document.getElementById("pdf-frame");
const $metadataPanel = document.getElementById("metadata-panel");
const $metaHeader = document.getElementById("meta-header");
const $metaAbstract = document.getElementById("meta-abstract");
const $metaSummary = document.getElementById("meta-summary");
const $metaNotes = document.getElementById("meta-notes");
const $btnBack = document.getElementById("btn-back");
const $sidebarOverlay = document.getElementById("sidebar-overlay");
const $detail = document.getElementById("detail");

// Fetch
async function fetchPapers() {
  const params = new URLSearchParams();
  if ($search.value) params.set("q", $search.value);
  if (selectedCategory) params.set("category", selectedCategory);
  const res = await fetch(`/api/papers?${params}`);
  const data = await res.json();
  papers = data.papers || [];
  categories = data.categories || [];
  renderSidebar();
  renderPaperList();
}

async function fetchDetail(id) {
  const res = await fetch(`/api/papers/${id}`);
  return await res.json();
}

// Sidebar
function renderSidebar() {
  const total = categories.reduce((s, c) => s + c.count, 0);
  let html = `<li class="${selectedCategory === null ? "active" : ""}" data-cat="">
    <span>All</span><span class="count">${total}</span></li>`;
  for (const c of categories) {
    html += `<li class="${selectedCategory === c.name ? "active" : ""}" data-cat="${c.name}">
      <span>${c.name}</span><span class="count">${c.count}</span></li>`;
  }
  $categories.innerHTML = html;
  $categories.querySelectorAll("li").forEach((li) => {
    li.onclick = () => {
      selectedCategory = li.dataset.cat || null;
      fetchPapers();
      writeURLState();
      closeMobileSidebar();
    };
  });
}

// Sort
function sortPapers(list) {
  const sorted = [...list];
  switch (sortOrder) {
    case "date-desc":
      return sorted.sort((a, b) => b.published.localeCompare(a.published));
    case "date-asc":
      return sorted.sort((a, b) => a.published.localeCompare(b.published));
    case "title":
      return sorted.sort((a, b) =>
        (a.title_ja || a.title).localeCompare(b.title_ja || b.title)
      );
    case "added":
      return sorted;
    default:
      return sorted;
  }
}

function formatAuthors(authors) {
  const filtered = (authors || []).filter((a) => a && a !== ":");
  if (filtered.length <= 3) return filtered.join(", ");
  return filtered.slice(0, 3).join(", ") + " et al.";
}

// Paper list
function renderPaperList() {
  const sorted = sortPapers(papers);
  $paperCount.textContent = `${sorted.length} papers`;
  let html = "";
  for (const p of sorted) {
    const title = p.title_ja || p.title;
    const indicators = [
      p.has_pdf_ja ? "JA" : "",
      p.has_summary ? "Sum" : "",
      p.has_note ? "Note" : "",
    ]
      .filter(Boolean)
      .join("  ");

    const q = $search.value;
    html += `<li class="${selectedPaperID === p.id ? "active" : ""}" data-id="${p.id}">
      <div class="paper-title">${highlight(title, q)}</div>
      <div class="paper-authors">${highlight(formatAuthors(p.authors), q)}</div>
      <div class="paper-meta">
        ${p.published}${indicators ? `<span class="paper-indicators">${indicators}</span>` : ""}
      </div>
    </li>`;
  }
  $papers.innerHTML = html;
  $papers.querySelectorAll("li").forEach((li) => {
    li.onclick = () => selectPaper(li.dataset.id);
  });
}

// Select paper
async function selectPaper(id) {
  selectedPaperID = id;
  writeURLState();
  renderPaperList();

  const detail = await fetchDetail(id);

  $detailEmpty.classList.add("hidden");
  $detailContent.classList.remove("hidden");

  // Mobile: slide in detail view
  if (isMobile()) {
    $detail.classList.add("mobile-show");
  }

  $toolbarId.textContent = detail.id;
  $btnArxiv.onclick = () =>
    window.open(`https://arxiv.org/abs/${detail.id}`, "_blank");

  $btnDownload.onclick = () => downloadPDF();

  // Language toggle
  showJapanese = false;
  if (detail.has_pdf_ja) {
    $langToggle.classList.remove("hidden");
    $btnEn.classList.add("active");
    $btnJa.classList.remove("active");
  } else {
    $langToggle.classList.add("hidden");
  }

  renderMetadata(detail);
  loadPDF(detail.id, false);
}

// Metadata
function renderMetadata(detail) {
  const titleDisplay = detail.title_ja || detail.title;
  const titleOrig =
    detail.title_ja && detail.title !== detail.title_ja
      ? `<div class="meta-title-orig">${esc(detail.title)}</div>`
      : "";
  const authors = (detail.authors || [])
    .filter((a) => a && a !== ":")
    .join(", ");
  const badges = [
    detail.has_pdf_ja ? "Translation" : "",
    detail.has_summary ? "Summary" : "",
    detail.has_note ? "Notes" : "",
  ].filter(Boolean);

  $metaHeader.innerHTML = `
    <div class="meta-title">${esc(titleDisplay)}</div>
    ${titleOrig}
    <div class="meta-authors">${esc(authors)}</div>
    <div class="meta-info">${detail.published}  ·  ${detail.category}  ·  ${detail.id}</div>
    ${badges.length ? `<div class="meta-badges">${badges.map((b) => `<span class="meta-badge">${b}</span>`).join("")}</div>` : ""}
  `;

  const abstract = detail.abstract_ja || detail.abstract;
  $metaAbstract.innerHTML = `
    <div class="meta-section">
      <div class="meta-section-title">Abstract</div>
      <div class="meta-section-body serif">${esc(abstract)}</div>
    </div>`;

  if (detail.has_summary) {
    fetch(`/api/papers/${detail.id}/summary`)
      .then((r) => r.text())
      .then((md) => {
        const html = marked.parse(md);
        const rewritten = html.replace(
          /src="assets\//g,
          `src="/api/papers/${detail.id}/assets/`
        );
        $metaSummary.innerHTML = `
          <div class="meta-section">
            <div class="meta-section-title">Summary</div>
            <div class="meta-section-body">${rewritten}</div>
          </div>`;
      });
  } else {
    $metaSummary.innerHTML = "";
  }

  if (detail.has_note) {
    fetch(`/api/papers/${detail.id}/note`)
      .then((r) => r.text())
      .then((md) => {
        $metaNotes.innerHTML = `
          <div class="meta-section">
            <div class="meta-section-title">Notes</div>
            <div class="meta-section-body">${marked.parse(md)}</div>
          </div>`;
      });
  } else {
    $metaNotes.innerHTML = "";
  }
}

// PDF — use browser's native PDF viewer via iframe (or direct link on mobile)
const $pdfLoading = document.getElementById("pdf-loading");
const $pdfMobilePrompt = document.getElementById("pdf-mobile-prompt");
const $pdfOpenLink = document.getElementById("pdf-open-link");

function loadPDF(id, japanese) {
  const url = japanese ? `/api/papers/${id}/pdf/ja` : `/api/papers/${id}/pdf`;
  if (isMobile()) {
    // Mobile: show a button to open PDF in a new tab (iframe PDF is unreliable on mobile)
    $pdfLoading.classList.add("hidden");
    $pdfMobilePrompt.classList.remove("hidden");
    $pdfOpenLink.href = url;
    $pdfFrame.src = "";
  } else {
    $pdfMobilePrompt.classList.add("hidden");
    $pdfLoading.classList.remove("hidden");
    $pdfFrame.onload = () => $pdfLoading.classList.add("hidden");
    $pdfFrame.src = url + "#view=FitH";
  }
}

// Download PDF
function downloadPDF() {
  if (!selectedPaperID) return;
  const base = showJapanese
    ? `/api/papers/${selectedPaperID}/pdf/ja`
    : `/api/papers/${selectedPaperID}/pdf`;
  window.open(base + "?download=1", "_blank");
}

// Events: language toggle
$btnEn.onclick = () => {
  if (!showJapanese) return;
  showJapanese = false;
  $btnEn.classList.add("active");
  $btnJa.classList.remove("active");
  if (selectedPaperID) loadPDF(selectedPaperID, false);
};

$btnJa.onclick = () => {
  if (showJapanese) return;
  showJapanese = true;
  $btnJa.classList.add("active");
  $btnEn.classList.remove("active");
  if (selectedPaperID) loadPDF(selectedPaperID, true);
};

// Events: theme toggle
const $btnToggleTheme = document.getElementById("btn-toggle-theme");

function applyTheme(theme) {
  if (theme) {
    document.documentElement.setAttribute("data-theme", theme);
  } else {
    document.documentElement.removeAttribute("data-theme");
  }
}

$btnToggleTheme.onclick = () => {
  const current = document.documentElement.getAttribute("data-theme");
  const isDark = current === "dark" ||
    (!current && window.matchMedia("(prefers-color-scheme: dark)").matches);
  const next = isDark ? "light" : "dark";
  applyTheme(next);
  localStorage.setItem("arq-theme", next);
};

// Restore saved theme
const savedTheme = localStorage.getItem("arq-theme");
if (savedTheme) applyTheme(savedTheme);

// Events: meta panel tabs
const $btnToggleNote = document.getElementById("btn-toggle-note");
const $metaTabInfo = document.getElementById("meta-tab-info");
const $metaTabNote = document.getElementById("meta-tab-note");
const $metaInfoContent = document.getElementById("meta-info-content");
const $metaNoteContent = document.getElementById("meta-note-content");

function switchMetaTab(tab) {
  if (tab === "info") {
    $metaTabInfo.classList.add("active");
    $metaTabNote.classList.remove("active");
    $metaInfoContent.classList.remove("hidden");
    $metaNoteContent.classList.add("hidden");
  } else {
    $metaTabNote.classList.add("active");
    $metaTabInfo.classList.remove("active");
    $metaNoteContent.classList.remove("hidden");
    $metaInfoContent.classList.add("hidden");
  }
}

$metaTabInfo.onclick = () => switchMetaTab("info");
$metaTabNote.onclick = () => switchMetaTab("note");

// Events: panel toggles
$btnToggleMeta.onclick = () => {
  if (!$metadataPanel.classList.contains("collapsed")) {
    // If note tab is active, just collapse
    $metadataPanel.classList.add("collapsed");
  } else {
    $metadataPanel.classList.remove("collapsed");
    switchMetaTab("info");
  }
};

$btnToggleNote.onclick = () => {
  if (!$metadataPanel.classList.contains("collapsed")) {
    const noteActive = $metaTabNote.classList.contains("active");
    if (noteActive) {
      $metadataPanel.classList.add("collapsed");
    } else {
      switchMetaTab("note");
    }
  } else {
    $metadataPanel.classList.remove("collapsed");
    switchMetaTab("note");
  }
};

$btnToggleList.onclick = () => {
  $paperList.classList.toggle("collapsed");
};

// Events: search
let searchTimer;
$search.oninput = () => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(() => {
    fetchPapers();
    writeURLState();
  }, 200);
};

// Events: sort
$sortSelect.onchange = () => {
  sortOrder = $sortSelect.value;
  renderPaperList();
  writeURLState();
};

// Events: keyboard
document.addEventListener("keydown", (e) => {
  if (e.target === $search) return;
  if ((e.key === "j" || e.key === "k") && !e.metaKey) {
    const sorted = sortPapers(papers);
    if (sorted.length === 0) return;
    const curIdx = sorted.findIndex((p) => p.id === selectedPaperID);
    let nextIdx;
    if (e.key === "j") {
      nextIdx = curIdx < 0 ? 0 : Math.min(curIdx + 1, sorted.length - 1);
    } else {
      nextIdx = curIdx < 0 ? 0 : Math.max(curIdx - 1, 0);
    }
    selectPaper(sorted[nextIdx].id);
    // Scroll the active item into view
    const activeLi = $papers.querySelector("li.active");
    if (activeLi) activeLi.scrollIntoView({ block: "nearest" });
  }
  if (e.key === "l" && !e.metaKey) {
    if ($langToggle.classList.contains("hidden")) return;
    showJapanese ? $btnEn.click() : $btnJa.click();
  }
  if (e.key === "i" && !e.metaKey) {
    $btnToggleMeta.click();
  }
  if (e.key === "n" && !e.metaKey) {
    $btnToggleNote.click();
  }
  if (e.key === "b" && !e.metaKey) {
    $btnToggleSidebar.click();
  }
  if (e.key === "t" && !e.metaKey) {
    $btnToggleTheme.click();
  }
  if (e.key === "d" && !e.metaKey) {
    downloadPDF();
  }
  if (e.key === "f" && !e.metaKey) {
    const isFullscreen = $sidebar.classList.contains("collapsed") &&
      $paperList.classList.contains("collapsed") &&
      $metadataPanel.classList.contains("collapsed");
    if (isFullscreen) {
      $sidebar.classList.remove("collapsed");
      $paperList.classList.remove("collapsed");
      $metadataPanel.classList.remove("collapsed");
    } else {
      $sidebar.classList.add("collapsed");
      $paperList.classList.add("collapsed");
      $metadataPanel.classList.add("collapsed");
    }
  }
});

function esc(s) {
  const d = document.createElement("div");
  d.textContent = s || "";
  return d.innerHTML;
}

function highlight(text, query) {
  if (!query) return esc(text);
  const escaped = esc(text);
  const qEsc = query.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  return escaped.replace(new RegExp(`(${qEsc})`, "gi"), `<mark>$1</mark>`);
}

// Mobile: FAB menu toggle
const $fabTrigger = document.getElementById("fab-trigger");
const $fabOverlay = document.getElementById("fab-overlay");
const $toolbarActions = document.getElementById("toolbar-actions");

function closeFab() {
  $toolbarActions.classList.remove("fab-open");
  $fabTrigger.classList.remove("fab-open");
  $fabOverlay.classList.remove("visible");
}

$fabTrigger.onclick = () => {
  const isOpen = $toolbarActions.classList.contains("fab-open");
  if (isOpen) {
    closeFab();
  } else {
    $toolbarActions.classList.add("fab-open");
    $fabTrigger.classList.add("fab-open");
    $fabOverlay.classList.add("visible");
  }
};

$fabOverlay.onclick = () => closeFab();

// Close FAB after any action button is tapped
$toolbarActions.addEventListener("click", (e) => {
  if (e.target.tagName === "BUTTON" && isMobile()) {
    closeFab();
  }
});

// Mobile: menu button in paper list header
const $btnMobileMenu = document.getElementById("btn-mobile-menu");
$btnMobileMenu.onclick = () => {
  $sidebar.classList.add("mobile-open");
  $sidebar.classList.remove("collapsed");
  $sidebarOverlay.classList.add("visible");
};

// Mobile: close sidebar when category is selected
function closeMobileSidebar() {
  if (isMobile()) {
    $sidebar.classList.remove("mobile-open");
    $sidebarOverlay.classList.remove("visible");
  }
}

// Mobile: back button returns to paper list
$btnBack.onclick = () => {
  $detail.classList.remove("mobile-show");
  // Collapse metadata panel if open
  $metadataPanel.classList.add("collapsed");
};

// Mobile: sidebar toggle (reuse hamburger button on list view)
$btnToggleSidebar.onclick = () => {
  if (isMobile()) {
    $sidebar.classList.toggle("mobile-open");
    $sidebar.classList.remove("collapsed");
    $sidebarOverlay.classList.toggle("visible");
  } else {
    $sidebar.classList.toggle("collapsed");
  }
};

// Mobile: close sidebar on overlay tap
$sidebarOverlay.onclick = () => {
  $sidebar.classList.remove("mobile-open");
  $sidebarOverlay.classList.remove("visible");
};

// Update check
const $updateBadge = document.getElementById("update-badge");

async function checkForUpdate() {
  try {
    const res = await fetch("/api/version");
    const data = await res.json();
    if (data.update_available && data.latest) {
      $updateBadge.textContent = `v${data.latest} available`;
      $updateBadge.classList.remove("hidden");
      $updateBadge.onclick = async () => {
        if (!confirm(`Current: v${data.current}\nLatest: v${data.latest}\n\nRun "brew upgrade arq" first, then click OK to restart the server.`)) return;
        $updateBadge.textContent = "restarting…";
        try {
          await fetch("/api/restart", { method: "POST" });
        } catch (_) {
          // Connection will drop as server restarts
        }
        setTimeout(() => location.reload(), 2000);
      };
    }
  } catch (_) {
    // Ignore errors (offline, etc.)
  }
}

// Init
async function init() {
  const state = readURLState();
  selectedCategory = state.category;
  sortOrder = state.sort;
  $sortSelect.value = sortOrder;
  $search.value = state.q;

  await fetchPapers();

  if (state.paper) {
    selectPaper(state.paper);
  }

  checkForUpdate();
}

init();

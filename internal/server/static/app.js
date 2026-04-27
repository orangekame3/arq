// State
let papers = [];
let categories = [];
let selectedCategory = null;
let selectedPaperID = null;
let showJapanese = false;
let sortOrder = "date-desc";

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
const $btnToggleMeta = document.getElementById("btn-toggle-meta");
const $btnToggleSidebar = document.getElementById("btn-toggle-sidebar");
const $btnToggleList = document.getElementById("btn-toggle-list");
const $pdfFrame = document.getElementById("pdf-frame");
const $metadataPanel = document.getElementById("metadata-panel");
const $metaHeader = document.getElementById("meta-header");
const $metaAbstract = document.getElementById("meta-abstract");
const $metaSummary = document.getElementById("meta-summary");
const $metaNotes = document.getElementById("meta-notes");

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

    html += `<li class="${selectedPaperID === p.id ? "active" : ""}" data-id="${p.id}">
      <div class="paper-title">${esc(title)}</div>
      <div class="paper-authors">${esc(formatAuthors(p.authors))}</div>
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
  window.location.hash = `paper=${id}`;
  renderPaperList();

  const detail = await fetchDetail(id);

  $detailEmpty.classList.add("hidden");
  $detailContent.classList.remove("hidden");

  $toolbarId.textContent = detail.id;
  $btnArxiv.onclick = () =>
    window.open(`https://arxiv.org/abs/${detail.id}`, "_blank");

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

// PDF — use browser's native PDF viewer via iframe
function loadPDF(id, japanese) {
  const url = japanese ? `/api/papers/${id}/pdf/ja` : `/api/papers/${id}/pdf`;
  $pdfFrame.src = url;
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

// Events: panel toggles
$btnToggleMeta.onclick = () => {
  $metadataPanel.classList.toggle("collapsed");
};

$btnToggleSidebar.onclick = () => {
  $sidebar.classList.toggle("collapsed");
};

$btnToggleList.onclick = () => {
  $paperList.classList.toggle("collapsed");
};

// Events: search
let searchTimer;
$search.oninput = () => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(fetchPapers, 200);
};

// Events: sort
$sortSelect.onchange = () => {
  sortOrder = $sortSelect.value;
  renderPaperList();
};

// Events: keyboard
document.addEventListener("keydown", (e) => {
  if (e.target === $search) return;
  if (e.key === "l" && !e.metaKey) {
    if ($langToggle.classList.contains("hidden")) return;
    showJapanese ? $btnEn.click() : $btnJa.click();
  }
  if (e.key === "i" && !e.metaKey) {
    $btnToggleMeta.click();
  }
  if (e.key === "b" && !e.metaKey) {
    $btnToggleSidebar.click();
  }
});

function esc(s) {
  const d = document.createElement("div");
  d.textContent = s || "";
  return d.innerHTML;
}

// Init
async function init() {
  await fetchPapers();
  const hash = window.location.hash;
  if (hash.startsWith("#paper=")) {
    selectPaper(hash.slice(7));
  }
}

init();

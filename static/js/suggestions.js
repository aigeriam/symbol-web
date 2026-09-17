// Text Suggestions feature: as the user types, debounce and ask
// /api/suggest for creative completions, then show them in a dropdown.
(function () {
  const input = document.getElementById("text-input");
  const dropdown = document.getElementById("suggestions-dropdown");
  if (!input || !dropdown) return;

  let debounceTimer;
  const DEBOUNCE_MS = 300;
  const MIN_LENGTH = 3;

  async function getSuggestions(text) {
    showLoading();
    try {
      const response = await fetch("/api/suggest", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text }),
      });

      if (!response.ok) {
        const err = await safeJson(response);
        showError(err?.error || "Could not get suggestions right now.");
        return;
      }

      const data = await response.json();
      displaySuggestions(data.suggestions || []);
    } catch (e) {
      showError("Could not reach the suggestions service.");
    }
  }

  function displaySuggestions(suggestions) {
    dropdown.innerHTML = "";
    if (!suggestions.length) {
      hide();
      return;
    }
    suggestions.forEach((s) => {
      const item = document.createElement("div");
      item.className = "suggestion-item";
      item.textContent = s;
      item.addEventListener("click", () => {
        input.value = s;
        hide();
        input.focus();
      });
      dropdown.appendChild(item);
    });
    show();
  }

  function showLoading() {
    dropdown.innerHTML = '<div class="loading">Getting suggestions…</div>';
    show();
  }

  function showError(message) {
    dropdown.innerHTML = `<div class="loading">${escapeHtml(message)}</div>`;
    show();
  }

  function show() {
    dropdown.classList.remove("hidden");
  }
  function hide() {
    dropdown.classList.add("hidden");
  }

  async function safeJson(response) {
    try {
      return await response.json();
    } catch {
      return null;
    }
  }

  function escapeHtml(str) {
    const div = document.createElement("div");
    div.textContent = str;
    return div.innerHTML;
  }

  input.addEventListener("input", (e) => {
    clearTimeout(debounceTimer);
    const value = e.target.value.trim();
    if (value.length < MIN_LENGTH) {
      hide();
      return;
    }
    debounceTimer = setTimeout(() => getSuggestions(value), DEBOUNCE_MS);
  });

  // Hide the dropdown when clicking elsewhere.
  document.addEventListener("click", (e) => {
    if (e.target !== input && !dropdown.contains(e.target)) {
      hide();
    }
  });
})();

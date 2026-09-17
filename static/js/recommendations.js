// Banner Recommendation feature: rule-based, calls /api/recommend-banner
// and highlights the recommended banner radio button.
(function () {
  const btn = document.getElementById("ai-recommend-btn");
  const resultBox = document.getElementById("recommendation-result");
  const textInput = document.getElementById("text-input");
  if (!btn || !resultBox || !textInput) return;

  btn.addEventListener("click", async () => {
    const text = textInput.value.trim();
    if (!text) {
      showMessage("Type some text first, then ask for a recommendation.");
      return;
    }

    setLoading(true);
    try {
      const response = await fetch("/api/recommend-banner", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text }),
      });

      if (!response.ok) {
        const err = await safeJson(response);
        showMessage(err?.error || "Could not get a recommendation right now.");
        return;
      }

      const data = await response.json();
      render(data);
      highlightBanner(data.recommended);
    } catch (e) {
      showMessage("Could not reach the recommendation service.");
    } finally {
      setLoading(false);
    }
  });

  function render(data) {
    const alternatives = (data.alternatives || [])
      .map(
        (a) =>
          `<div class="rec-alt">${escapeHtml(a.banner)} (score ${a.score}) — ${escapeHtml(a.reason)}</div>`
      )
      .join("");

    resultBox.innerHTML = `
      <strong>Recommended: ${escapeHtml(data.recommended)}</strong>
      <div>${escapeHtml(data.reasoning)}</div>
      ${alternatives}
    `;
    resultBox.classList.remove("hidden");
  }

  function showMessage(msg) {
    resultBox.innerHTML = `<div>${escapeHtml(msg)}</div>`;
    resultBox.classList.remove("hidden");
  }

  function setLoading(isLoading) {
    btn.disabled = isLoading;
    btn.textContent = isLoading ? "🤖 Thinking…" : "🤖 AI Recommend Banner";
  }

  function highlightBanner(banner) {
    document.querySelectorAll('input[name="banner"]').forEach((radio) => {
      radio.checked = radio.value === banner;
    });
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
    div.textContent = String(str);
    return div.innerHTML;
  }
})();

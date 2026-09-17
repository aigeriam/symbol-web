// Creative Variations feature (bonus): calls /api/variations and shows the
// results in a modal; clicking one fills the text input and picks its
// suggested banner.
(function () {
  const openBtn = document.getElementById("get-variations-btn");
  const modal = document.getElementById("variations-modal");
  const closeBtn = document.getElementById("close-variations");
  const list = document.getElementById("variations-list");
  const textInput = document.getElementById("text-input");
  if (!openBtn || !modal || !closeBtn || !list || !textInput) return;

  openBtn.addEventListener("click", async () => {
    const text = textInput.value.trim();
    if (!text) {
      alert("Type some text first, then ask for ideas.");
      return;
    }

    list.innerHTML = '<div class="loading">Cooking up some ideas…</div>';
    modal.classList.remove("hidden");

    try {
      const response = await fetch("/api/variations", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text }),
      });

      if (!response.ok) {
        const err = await safeJson(response);
        list.innerHTML = `<div class="loading">${escapeHtml(
          err?.error || "Could not get variations right now."
        )}</div>`;
        return;
      }

      const data = await response.json();
      renderVariations(data.variations || []);
    } catch (e) {
      list.innerHTML = '<div class="loading">Could not reach the variations service.</div>';
    }
  });

  closeBtn.addEventListener("click", () => modal.classList.add("hidden"));
  modal.addEventListener("click", (e) => {
    if (e.target === modal) modal.classList.add("hidden");
  });

  function renderVariations(variations) {
    list.innerHTML = "";
    if (!variations.length) {
      list.innerHTML = '<div class="loading">No variations came back — try different text.</div>';
      return;
    }
    variations.forEach((v) => {
      const card = document.createElement("div");
      card.className = "variation-card";
      card.innerHTML = `
        <div><strong>${escapeHtml(v.text)}</strong></div>
        <div class="v-desc">${escapeHtml(v.description)} · banner: ${escapeHtml(v.suggested_banner)}</div>
      `;
      card.addEventListener("click", () => {
        textInput.value = v.text;
        document.querySelectorAll('input[name="banner"]').forEach((radio) => {
          radio.checked = radio.value === v.suggested_banner;
        });
        modal.classList.add("hidden");
      });
      list.appendChild(card);
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

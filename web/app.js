const API_URL = "http://localhost:8080/v1/cash-in/qr";

const form = document.getElementById("qr-form");
const referenceIdInput = document.getElementById("reference_id");
const amountInput = document.getElementById("amount");
const simulateBtn = document.getElementById("simulate-btn");
const statusEl = document.getElementById("status");
const canvas = document.getElementById("qr");

let lastRequestId = null;
let lastAmount = 0;
let pollTimer = null;

referenceIdInput.value = "test-" + Date.now();

form.addEventListener("submit", async (event) => {
  event.preventDefault();

  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }

  statusEl.textContent = "Generating...";

  try {
    const response = await fetch(API_URL, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        reference_id: referenceIdInput.value,
        amount: Number(amountInput.value),
      }),
    });

    const data = await response.json();

    if (!response.ok) {
      statusEl.textContent = data.error || "Request failed";
      return;
    }

    await QRCode.toCanvas(canvas, data.qr_string, { width: 280 });

    lastRequestId = data.payment_request_id;
    lastAmount = Number(amountInput.value);
    simulateBtn.hidden = false;

    statusEl.textContent = data.payment_request_id + " (" + data.status + ")";
  } catch (err) {
    statusEl.textContent = err.message;
  }
});

simulateBtn.addEventListener("click", async () => {
  if (!lastRequestId) {
    return;
  }

  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }

  statusEl.textContent = "Simulating payment...";

  try {
    const response = await fetch(API_URL + "/" + lastRequestId + "/simulate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ amount: lastAmount }),
    });

    const data = await response.json();

    if (!response.ok) {
      statusEl.textContent = data.error || "Simulation failed";
      return;
    }

    statusEl.textContent = "Simulation " + data.status + ". Checking status...";
    startPolling();
  } catch (err) {
    statusEl.textContent = err.message;
  }
});

function startPolling() {
  let attempts = 0;
  const maxAttempts = 15;

  pollTimer = setInterval(async () => {
    attempts++;

    try {
      const response = await fetch(API_URL + "/" + lastRequestId);
      const data = await response.json();

      if (!response.ok) {
        statusEl.textContent = data.error || "Status check failed";
        clearInterval(pollTimer);
        pollTimer = null;
        return;
      }

      statusEl.textContent = data.payment_request_id + " (" + data.status + ")";

      if (["SUCCEEDED", "FAILED", "EXPIRED", "CANCELED"].includes(data.status)) {
        clearInterval(pollTimer);
        pollTimer = null;
      } else if (attempts >= maxAttempts) {
        clearInterval(pollTimer);
        pollTimer = null;
        statusEl.textContent += " (timed out)";
      }
    } catch (err) {
      statusEl.textContent = err.message;
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }, 2000);
}

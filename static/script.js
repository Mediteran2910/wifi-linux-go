// --- Utility Function ---
function getSignalIcon(rssi) {
  if (rssi > 80) {
    return "/static/public/wifi-signal-5.svg"; // Excellent signal
  } else if (rssi > 60) {
    return "/static/public/wifi-signal-4.svg"; // Very good signal
  } else if (rssi > 40) {
    return "/static/public/wifi-signal-3.svg"; // Good signal
  } else if (rssi > 20) {
    return "/static/public/wifi-signal-2.svg"; // Fair signal
  } else {
    return "/static/public/wifi-signal-1.svg"; // Weak signal
  }
}

// --- DOM Element References ---
const mainAppContainer = document.getElementById("main-app");
const networksListContainer = document.getElementById("networks-list");
const modalOverlay = document.getElementById("wifi-connect-modal");
const modalCloseBtn = modalOverlay.querySelector(".modal-close-btn");
const networkNameEl = modalOverlay.querySelector(".network-name");
const passwordFieldContainer = modalOverlay.querySelector(".password-field-container");
const passwordInput = modalOverlay.querySelector("#password-input");
const togglePasswordBtn = modalOverlay.querySelector(".toggle-password-btn");
const connectForm = modalOverlay.querySelector("#connect-form");
const connectBtn = modalOverlay.querySelector("#connect-btn");
const connectionErrorEl = modalOverlay.querySelector("#connection-error");

const loadingMessageEl = document.getElementById("loading-message");
const errorMessageEl = document.getElementById("error-message");
// const connectingMessageEl = document.getElementById("connecting-message"); // Commented out
const successPageContainer = document.getElementById("connection-success-page");
const connectedNetworkNameEl = document.getElementById("connected-network-name");

// --- State (Plain JS) ---
let currentNetwork = null;

// --- Functions to Update the UI ---

function setStatus(loading, error, connecting, networksFound) {
    loadingMessageEl.classList.toggle("hidden", !loading);
    errorMessageEl.classList.toggle("hidden", !error);
    // connectingMessageEl.classList.toggle("hidden", !connecting); // Commented out
    networksListContainer.classList.toggle("hidden", !networksFound);
    mainAppContainer.classList.toggle("hidden", false);
}

// function showSuccessPage(ssid) { // Commented out
//     mainAppContainer.classList.add("hidden");
//     successPageContainer.classList.remove("hidden");
//     connectedNetworkNameEl.textContent = ssid;
// }

function showModal(network) {
    currentNetwork = network;
    networkNameEl.textContent = network.name;
    passwordInput.value = "";
    connectionErrorEl.classList.add("hidden");
    
    const showPasswordField = network.security && network.security.toLowerCase() !== "none";
    passwordFieldContainer.classList.toggle("hidden", !showPasswordField);

    modalOverlay.classList.remove("hidden");
    
    setStatus(false, false, false, true);
}

function hideModal() {
    modalOverlay.classList.add("hidden");
}

function renderNetworks(networks) {
    networksListContainer.innerHTML = "";
    if (!networks || networks.length === 0) {
        setStatus(false, false, false, false);
        loadingMessageEl.textContent = "No Wi-Fi networks found.";
        loadingMessageEl.classList.remove("hidden");
        return;
    }

    const ul = document.createElement("ul");
    networks.forEach((network) => {
        const li = document.createElement("li");
        li.className = "wifi-list-item";
        li.addEventListener("click", () => handleNetworkSelect(network));
        
        const signalImagePath = getSignalIcon(network.rssi);

        li.innerHTML = `
            <span class="network-name-text">${network.name}</span>
            <div class="network-icons">
                <img src="/static/public/${network.security.toLowerCase() !== "none" ? "locked.svg" : "unlocked.svg"}" alt="Security icon" class="security-icon">
                <img src="${signalImagePath}" alt="Signal icon" class="wifi-icon">
            </div>
        `;
        ul.appendChild(li);
    });
    networksListContainer.appendChild(ul);
    setStatus(false, false, false, true);
}

// --- Event Handlers ---
async function handleNetworkSelect(network) {
    currentNetwork = network;
    // setStatus(false, false, true, false); // Commented out
    // connectingMessageEl.textContent = `Attempting to connect to ${network.name || 'network'}...`; // Commented out

    // Directly show the modal and do not attempt to connect automatically.
    // The password prompt will handle whether a password is required.
    showModal(network);
    setStatus(false, false, false, true); // Reset status to show networks list
}

modalCloseBtn.addEventListener("click", hideModal);

togglePasswordBtn.addEventListener("click", () => {
  const isPassword = passwordInput.type === "password";
  passwordInput.type = isPassword ? "text" : "password";
  const icon = togglePasswordBtn.querySelector("img");
  icon.src = isPassword
    ? "/static/public/hidden-pass.svg"
    : "/static/public/show-pass.svg";
  icon.alt = isPassword ? "Hide password icon" : "Show password icon";
});

connectForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  connectBtn.textContent = "Connecting...";
  connectBtn.disabled = true;
  passwordInput.disabled = true;
  connectionErrorEl.classList.add("hidden");

  const ssid = currentNetwork.name;
  const password = passwordInput.value;

  try {
    const response = await fetch("/api/wifi/connect", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ssid, password }),
    });

    const data = await response.json();

    if (response.ok) {
      console.log("Connection successful:", data.message);
      connectionErrorEl.textContent = "Successfully connected! Your device will now be disconnected from the captive portal. Please check if your router has changed color.";
      connectionErrorEl.style.color = "#4ade80"; // Set color to green for success
      connectionErrorEl.classList.remove("hidden");
    } else {
      console.error(
        "Connection failed:",
        data.error || data.details || "Unknown error"
      );
      connectionErrorEl.textContent = data.error || "Failed to connect. Please try again.";
      connectionErrorEl.classList.remove("hidden");
      
      if (response.status === 401 && data.error && data.error.includes("Authentication failed")) {
          passwordInput.value = "";
      }
    }
  } catch (error) {
    console.error("Network request error:", error);
    connectionErrorEl.textContent = "Could not reach the server.";
    connectionErrorEl.classList.remove("hidden");
  } finally {
    connectBtn.textContent = "Connect";
    connectBtn.disabled = false;
    passwordInput.disabled = false;
  }
});

// --- Initial Data Fetch ---
async function fetchNetworks() {
  setStatus(true, false, false, false);
  loadingMessageEl.textContent = "Searching for Wi-Fi networks...";
  try {
    const response = await fetch("/api/wifi/scan");
    if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }
    const networks = await response.json();
    networks.sort((a, b) => (b.rssi || 0) - (a.rssi || 0));
    renderNetworks(networks);
  } catch (e) {
    console.error("Failed to fetch Wi-Fi networks:", e);
    setStatus(false, true, false, false);
    errorMessageEl.textContent = `Error: ${e.message}`;
  }
}

fetchNetworks();
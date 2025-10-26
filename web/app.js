// State
let currentConversation = null;
let conversations = [];
let messages = [];
let refreshInterval = null;

// DOM Elements
const conversationsList = document.getElementById('conversationsList');
const emptyState = document.getElementById('emptyState');
const chatArea = document.getElementById('chatArea');
const chatUserName = document.getElementById('chatUserName');
const chatUsername = document.getElementById('chatUsername');
const messagesContainer = document.getElementById('messagesContainer');
const messageInput = document.getElementById('messageInput');
const sendBtn = document.getElementById('sendBtn');
const toggleBotBtn = document.getElementById('toggleBotBtn');
const settingsBtn = document.getElementById('settingsBtn');
const settingsModal = document.getElementById('settingsModal');
const closeBtn = document.querySelector('.close-btn');
const cancelBtn = document.getElementById('cancelBtn');
const saveKBBtn = document.getElementById('saveKBBtn');
const knowledgeBaseInput = document.getElementById('knowledgeBaseInput');

// Initialize
init();

function init() {
    loadConversations();
    setupEventListeners();
    startAutoRefresh();
}

function setupEventListeners() {
    sendBtn.addEventListener('click', sendMessage);
    messageInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    });

    toggleBotBtn.addEventListener('click', toggleBot);
    settingsBtn.addEventListener('click', openSettings);
    closeBtn.addEventListener('click', closeSettings);
    cancelBtn.addEventListener('click', closeSettings);
    saveKBBtn.addEventListener('click', saveKnowledgeBase);

    // Close modal on outside click
    settingsModal.addEventListener('click', (e) => {
        if (e.target === settingsModal) {
            closeSettings();
        }
    });
}

// Auto refresh
function startAutoRefresh() {
    refreshInterval = setInterval(() => {
        loadConversations();
        if (currentConversation) {
            loadMessages(currentConversation.id);
        }
    }, 3000);
}

// API Calls
async function loadConversations() {
    try {
        const response = await fetch('/api/conversations');
        const data = await response.json();
        conversations = data || [];
        renderConversations();
    } catch (error) {
        console.error('Error loading conversations:', error);
    }
}

async function loadMessages(conversationId) {
    try {
        const response = await fetch(`/api/conversations/${conversationId}/messages`);
        const data = await response.json();
        messages = data || [];
        renderMessages();
        scrollToBottom();
    } catch (error) {
        console.error('Error loading messages:', error);
    }
}

async function sendMessage() {
    if (!currentConversation) return;

    const text = messageInput.value.trim();
    if (!text) return;

    try {
        const response = await fetch(`/api/conversations/${currentConversation.id}/send`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ message: text }),
        });

        if (response.ok) {
            messageInput.value = '';
            loadMessages(currentConversation.id);
        } else {
            alert('Failed to send message');
        }
    } catch (error) {
        console.error('Error sending message:', error);
        alert('Error sending message');
    }
}

async function toggleBot() {
    if (!currentConversation) return;

    const endpoint = currentConversation.is_bot_active ? 'takeover' : 'activate-bot';

    try {
        const response = await fetch(`/api/conversations/${currentConversation.id}/${endpoint}`, {
            method: 'POST',
        });

        if (response.ok) {
            currentConversation.is_bot_active = !currentConversation.is_bot_active;
            updateToggleButton();
            loadConversations();
        } else {
            alert('Failed to toggle bot');
        }
    } catch (error) {
        console.error('Error toggling bot:', error);
        alert('Error toggling bot');
    }
}

async function openSettings() {
    try {
        const response = await fetch('/api/knowledge-base');
        const data = await response.json();
        knowledgeBaseInput.value = data.content || '';
        settingsModal.classList.add('active');
    } catch (error) {
        console.error('Error loading knowledge base:', error);
        alert('Error loading knowledge base');
    }
}

function closeSettings() {
    settingsModal.classList.remove('active');
}

async function saveKnowledgeBase() {
    const content = knowledgeBaseInput.value.trim();
    if (!content) {
        alert('Knowledge base cannot be empty');
        return;
    }

    try {
        const response = await fetch('/api/knowledge-base', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ content }),
        });

        if (response.ok) {
            alert('Knowledge base updated successfully');
            closeSettings();
        } else {
            alert('Failed to update knowledge base');
        }
    } catch (error) {
        console.error('Error updating knowledge base:', error);
        alert('Error updating knowledge base');
    }
}

// Rendering
function renderConversations() {
    if (conversations.length === 0) {
        conversationsList.innerHTML = '<div class="loading">No conversations yet</div>';
        return;
    }

    conversationsList.innerHTML = conversations.map(conv => {
        const isActive = currentConversation && currentConversation.id === conv.id;
        const statusLabel = conv.is_bot_active ? 'Bot Active' : 'Admin Mode';
        const statusClass = conv.is_bot_active ? '' : 'inactive';
        const displayName = conv.telegram_first_name || conv.telegram_username || 'User';
        const username = conv.telegram_username ? `@${conv.telegram_username}` : `ID: ${conv.telegram_chat_id}`;

        return `
            <div class="conversation-item ${isActive ? 'active' : ''}" data-id="${conv.id}">
                <div class="conversation-name">
                    ${displayName}
                    <span class="bot-status ${statusClass}">${statusLabel}</span>
                </div>
                <div class="conversation-preview">${conv.last_message || 'No messages yet'}</div>
                <div class="conversation-time">${formatTime(conv.last_message_time || conv.created_at)}</div>
            </div>
        `;
    }).join('');

    // Add click handlers
    document.querySelectorAll('.conversation-item').forEach(item => {
        item.addEventListener('click', () => {
            const id = parseInt(item.dataset.id);
            selectConversation(id);
        });
    });
}

function selectConversation(id) {
    currentConversation = conversations.find(c => c.id === id);
    if (!currentConversation) return;

    emptyState.style.display = 'none';
    chatArea.style.display = 'flex';

    const displayName = currentConversation.telegram_first_name || currentConversation.telegram_username || 'User';
    const username = currentConversation.telegram_username ? `@${currentConversation.telegram_username}` : `ID: ${currentConversation.telegram_chat_id}`;

    chatUserName.textContent = displayName;
    chatUsername.textContent = username;

    updateToggleButton();
    loadMessages(id);
    renderConversations(); // Re-render to update active state
}

function updateToggleButton() {
    if (currentConversation.is_bot_active) {
        toggleBotBtn.textContent = 'Take Over';
        toggleBotBtn.className = 'btn btn-primary';
    } else {
        toggleBotBtn.textContent = 'Activate Bot';
        toggleBotBtn.className = 'btn btn-secondary';
    }
}

function renderMessages() {
    if (messages.length === 0) {
        messagesContainer.innerHTML = '<div class="loading">No messages yet</div>';
        return;
    }

    messagesContainer.innerHTML = messages.map(msg => {
        const senderLabel = msg.sender_type === 'admin' ? 'Admin' : msg.sender_type === 'bot' ? 'Bot' : '';

        return `
            <div class="message ${msg.sender_type}">
                ${msg.sender_type !== 'user' ? `<div class="message-sender">${senderLabel}</div>` : ''}
                <div class="message-text">${escapeHtml(msg.message_text)}</div>
                <div class="message-time">${formatTime(msg.created_at)}</div>
            </div>
        `;
    }).join('');
}

// Helpers
function formatTime(dateStr) {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;

    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;

    const diffDays = Math.floor(diffHours / 24);
    if (diffDays < 7) return `${diffDays}d ago`;

    return date.toLocaleDateString();
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function scrollToBottom() {
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

// Notification System JavaScript

let notificationDropdownOpen = false;

document.addEventListener('DOMContentLoaded', function() {
    const notificationBtn = document.getElementById('notificationBtn');
    const notificationDropdown = document.getElementById('notificationDropdown');
    const markAllReadBtn = document.getElementById('markAllReadBtn');

    if (!notificationBtn) return; // Not logged in

    // Toggle dropdown
    notificationBtn.addEventListener('click', function(e) {
        e.stopPropagation();
        notificationDropdownOpen = !notificationDropdownOpen;
        
        if (notificationDropdownOpen) {
            notificationDropdown.style.display = 'block';
            loadNotifications();
        } else {
            notificationDropdown.style.display = 'none';
        }
    });

    // Close dropdown when clicking outside
    document.addEventListener('click', function(e) {
        if (notificationDropdownOpen && !notificationDropdown.contains(e.target) && e.target !== notificationBtn) {
            notificationDropdown.style.display = 'none';
            notificationDropdownOpen = false;
        }
    });

    // Mark all as read
    markAllReadBtn.addEventListener('click', function() {
        markAllNotificationsAsRead();
    });

    // Load unread count on page load
    updateUnreadCount();

    // Auto-refresh unread count every 30 seconds
    setInterval(updateUnreadCount, 30000);
});

// Load notifications
function loadNotifications() {
    fetch('/api/notifications?limit=10')
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                displayNotifications(data.notifications);
            }
        })
        .catch(error => {
            console.error('Error loading notifications:', error);
        });
}

// Display notifications in dropdown
function displayNotifications(notifications) {
    const notificationsList = document.getElementById('notificationsList');
    
    if (!notifications || notifications.length === 0) {
        notificationsList.innerHTML = `
            <div style="padding: 40px 20px; text-align: center; color: var(--text-muted);">
                <span class="material-icons" style="font-size: 48px; opacity: 0.5;">notifications_none</span>
                <p style="margin: 8px 0 0 0; font-size: 13px;">Tidak ada notifikasi</p>
            </div>
        `;
        return;
    }

    let html = '';
    notifications.forEach(notification => {
        const iconMap = {
            'report': 'report',
            'warning': 'warning',
            'system_update': 'info'
        };

        const colorMap = {
            'report': '#faa61a',
            'warning': '#dc3545',
            'system_update': '#007bff'
        };

        const icon = iconMap[notification.Type] || 'notifications';
        const color = colorMap[notification.Type] || 'var(--text-muted)';
        const unreadClass = notification.IsRead ? '' : 'background: rgba(250, 166, 26, 0.05);';

        html += `
            <div class="notification-item" data-id="${notification.ID}" style="padding: 12px 16px; border-bottom: 1px solid var(--bg-tertiary); cursor: pointer; ${unreadClass}" onclick="handleNotificationClick(${notification.ID}, '${notification.ReferenceURL || ''}', ${notification.IsRead})">
                <div style="display: flex; gap: 12px;">
                    <div style="flex-shrink: 0;">
                        <span class="material-icons" style="color: ${color}; font-size: 20px;">${icon}</span>
                    </div>
                    <div style="flex: 1;">
                        <div style="font-weight: ${notification.IsRead ? 'normal' : '600'}; color: var(--text-header); font-size: 13px; margin-bottom: 4px;">
                            ${notification.Title}
                        </div>
                        <div style="color: var(--text-normal); font-size: 12px; margin-bottom: 4px;">
                            ${notification.Message}
                        </div>
                        <div style="color: var(--text-muted); font-size: 11px;">
                            ${formatNotificationTime(notification.CreatedAt)}
                        </div>
                    </div>
                    ${!notification.IsRead ? '<div style="width: 8px; height: 8px; background: #faa61a; border-radius: 50%; flex-shrink: 0; margin-top: 6px;"></div>' : ''}
                </div>
            </div>
        `;
    });

    notificationsList.innerHTML = html;
}

// Handle notification click
function handleNotificationClick(notificationId, referenceUrl, isRead) {
    // Mark as read if not already
    if (!isRead) {
        markNotificationAsRead(notificationId);
    }

    // Navigate to reference URL if exists
    if (referenceUrl) {
        window.location.href = referenceUrl;
    }
}

// Mark notification as read
function markNotificationAsRead(notificationId) {
    fetch(`/api/notifications/${notificationId}/read`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            updateUnreadCount();
        }
    })
    .catch(error => {
        console.error('Error marking notification as read:', error);
    });
}

// Mark all notifications as read
function markAllNotificationsAsRead() {
    fetch('/api/notifications/read-all', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            updateUnreadCount();
            loadNotifications(); // Reload to update UI
        }
    })
    .catch(error => {
        console.error('Error marking all as read:', error);
    });
}

// Update unread count badge
function updateUnreadCount() {
    fetch('/api/notifications/count')
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                const badge = document.getElementById('notificationBadge');
                if (badge) {
                    if (data.count > 0) {
                        badge.textContent = data.count > 99 ? '99+' : data.count;
                        badge.style.display = 'flex';
                    } else {
                        badge.style.display = 'none';
                    }
                }
            }
        })
        .catch(error => {
            console.error('Error updating unread count:', error);
        });
}

// Format notification time
function formatNotificationTime(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diffInSeconds = Math.floor((now - date) / 1000);

    if (diffInSeconds < 60) {
        return 'Baru saja';
    } else if (diffInSeconds < 3600) {
        const minutes = Math.floor(diffInSeconds / 60);
        return `${minutes} menit yang lalu`;
    } else if (diffInSeconds < 86400) {
        const hours = Math.floor(diffInSeconds / 3600);
        return `${hours} jam yang lalu`;
    } else if (diffInSeconds < 604800) {
        const days = Math.floor(diffInSeconds / 86400);
        return `${days} hari yang lalu`;
    } else {
        return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
    }
}

// Report Post Functionality
function showReportModal(itemId) {
    const modal = document.getElementById('reportModal');
    if (modal) {
        modal.style.display = 'flex';
        document.getElementById('reportItemId').value = itemId;
    }
}

function closeReportModal() {
    const modal = document.getElementById('reportModal');
    if (modal) {
        modal.style.display = 'none';
        document.getElementById('reportForm').reset();
    }
}

function submitReport(event) {
    event.preventDefault();
    
    const form = event.target;
    const formData = new FormData(form);
    const itemId = document.getElementById('reportItemId').value;

    fetch(`/item/${itemId}/report`, {
        method: 'POST',
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            alert('✅ ' + data.message);
            closeReportModal();
        } else {
            alert('❌ ' + (data.error || 'Gagal mengirim laporan'));
        }
    })
    .catch(error => {
        console.error('Error submitting report:', error);
        alert('❌ Terjadi kesalahan saat mengirim laporan');
    });
}

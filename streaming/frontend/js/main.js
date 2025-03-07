class FileDownloader {
    constructor() {
        this.API_BASE_URL = '/api';
        this.files = [];
        this.selectedFiles = new Set();

        // DOM Elements
        this.elements = {
            fileList: document.getElementById('fileListBody'),
            selectAll: document.getElementById('selectAll'),
            downloadBtn: document.getElementById('downloadSelected'),
            searchInput: document.getElementById('searchInput'),
            selectedCount: document.getElementById('selectedCount'),
            totalSize: document.getElementById('totalSize')
        };

        this.initializeEventListeners();
        this.loadFiles();
    }

    initializeEventListeners() {
        // Select all checkbox
        this.elements.selectAll.addEventListener('change', () => {
            const checkboxes = document.querySelectorAll('.file-checkbox');
            checkboxes.forEach(checkbox => {
                checkbox.checked = this.elements.selectAll.checked;
                this.handleFileSelection(checkbox);
            });
        });

        // Download button
        this.elements.downloadBtn.addEventListener('click', () => {
            this.downloadSelectedFiles();
        });

        // Search input
        this.elements.searchInput.addEventListener('input', () => {
            this.filterFiles(this.elements.searchInput.value);
        });
    }

    async loadFiles() {
        try {
            const response = await fetch(`${this.API_BASE_URL}/files`);
            this.files = await response.json();
            this.renderFiles(this.files);
        } catch (error) {
            console.error('Error loading files:', error);
            this.showError('Failed to load files');
        }
    }

    renderFiles(files) {
        this.elements.fileList.innerHTML = files.map(file => this.createFileElement(file)).join('');
        
        // Add event listeners to checkboxes
        document.querySelectorAll('.file-checkbox').forEach(checkbox => {
            checkbox.addEventListener('change', () => this.handleFileSelection(checkbox));
        });
    }

    createFileElement(file) {
        const icon = file.isDirectory ? 'fa-folder' : this.getFileIcon(file.name);
        return `
            <div class="file-item" data-path="${file.path}">
                <div class="checkbox-cell">
                    <input type="checkbox" class="file-checkbox" data-path="${file.path}" data-size="${file.size}">
                </div>
                <div class="name-cell">
                    <i class="fas ${icon}"></i>
                    ${file.name}
                </div>
                <div class="size-cell">${this.formatSize(file.size)}</div>
                <div class="date-cell">${this.formatDate(file.lastModified)}</div>
            </div>
        `;
    }

    handleFileSelection(checkbox) {
        if (checkbox.checked) {
            this.selectedFiles.add(checkbox.dataset.path);
        } else {
            this.selectedFiles.delete(checkbox.dataset.path);
        }

        this.updateUI();
    }

    updateUI() {
        const selectedCount = this.selectedFiles.size;
        this.elements.downloadBtn.disabled = selectedCount === 0;
        this.elements.selectedCount.textContent = `${selectedCount} file${selectedCount !== 1 ? 's' : ''} selected`;

        // Calculate total size
        const totalSize = Array.from(document.querySelectorAll('.file-checkbox:checked'))
            .reduce((sum, checkbox) => sum + parseInt(checkbox.dataset.size), 0);
        this.elements.totalSize.textContent = `Total: ${this.formatSize(totalSize)}`;

        // Update select all checkbox
        const checkboxes = document.querySelectorAll('.file-checkbox');
        this.elements.selectAll.checked = selectedCount > 0 && selectedCount === checkboxes.length;
        this.elements.selectAll.indeterminate = selectedCount > 0 && selectedCount < checkboxes.length;
    }

    filterFiles(searchTerm) {
        const filtered = this.files.filter(file => 
            file.name.toLowerCase().includes(searchTerm.toLowerCase())
        );
        this.renderFiles(filtered);
    }

    downloadSelectedFiles() {
        if (this.selectedFiles.size === 0) return;

        const queryString = Array.from(this.selectedFiles)
            .map(file => `files=${encodeURIComponent(file)}`)
            .join('&');

        // Create a hidden link and trigger the download
        const downloadUrl = `${this.API_BASE_URL}/download?${queryString}`;
        window.location.href = downloadUrl;
    }

    getFileIcon(filename) {
        const ext = filename.split('.').pop().toLowerCase();
        const icons = {
            pdf: 'fa-file-pdf',
            doc: 'fa-file-word',
            docx: 'fa-file-word',
            xls: 'fa-file-excel',
            xlsx: 'fa-file-excel',
            png: 'fa-file-image',
            jpg: 'fa-file-image',
            jpeg: 'fa-file-image',
            gif: 'fa-file-image',
            zip: 'fa-file-archive',
            rar: 'fa-file-archive',
            txt: 'fa-file-alt',
            mp3: 'fa-file-audio',
            mp4: 'fa-file-video'
        };
        return icons[ext] || 'fa-file';
    }

    formatSize(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    formatDate(timestamp) {
        const date = new Date(timestamp);
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
    }

    showError(message) {
        alert(message);
    }
}

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    new FileDownloader();
}); 
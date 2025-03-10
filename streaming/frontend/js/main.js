class FileDownloader {
    constructor() {
        this.API_BASE_URL = '/api';
        this.files = [];
        this.selectedFiles = new Set();
        this.selectedFolders = new Set(); // Track selected folders separately
        this.currentPath = '';  // Track current directory path

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

    async loadFiles(path = '') {
        try {
            // Build URL with path query parameter if provided
            const url = path ? 
                `${this.API_BASE_URL}/files?path=${encodeURIComponent(path)}` : 
                `${this.API_BASE_URL}/files`;
            
            const response = await fetch(url);
            const data = await response.json();
            
            // Clear selections when changing directories
            this.selectedFiles.clear();
            this.selectedFolders.clear();
            
            // Update properties with new API response structure
            this.files = data.files || [];
            this.currentPath = data.currentPath || '';
            
            // Add a parent directory entry if we're not at the root
            if (this.currentPath && this.currentPath !== '.') {
                this.files.unshift(this.createParentDirectoryEntry());
            }
            
            this.renderFiles(this.files);
            
            // Update page title with current path
            this.updatePathDisplay();
            
            // Update UI to reflect cleared selections
            this.updateUI();
        } catch (error) {
            console.error('Error loading files:', error);
            this.showError('Failed to load files');
        }
    }

    // Create a special entry for parent directory navigation
    createParentDirectoryEntry() {
        return {
            name: '..',
            size: 0,
            path: this.getParentPath(this.currentPath),
            lastModified: new Date(),
            isDirectory: true,
            isParentDir: true
        };
    }

    // Get parent directory path
    getParentPath(path) {
        if (!path || path === '.' || path === '/') return '';
        
        const parts = path.split('/');
        parts.pop(); // Remove last segment
        return parts.join('/') || '.';
    }

    // Update display with current path
    updatePathDisplay() {
        const pathDisplay = document.createElement('div');
        pathDisplay.className = 'path-display';
        
        // Create breadcrumb elements
        let pathSoFar = '';
        const pathParts = [];
        
        // Add root
        pathParts.push({
            text: 'Root',
            path: ''
        });
        
        // Add path segments
        if (this.currentPath && this.currentPath !== '.') {
            const segments = this.currentPath.split('/');
            segments.forEach((segment, index) => {
                pathSoFar += (index > 0 ? '/' : '') + segment;
                pathParts.push({
                    text: segment,
                    path: pathSoFar
                });
            });
        }
        
        // Create breadcrumb HTML
        const breadcrumbHtml = pathParts.map((part, index) => {
            const isLast = index === pathParts.length - 1;
            return isLast ? 
                `<span class="breadcrumb-current">${part.text}</span>` : 
                `<a href="#" class="breadcrumb-link" data-path="${part.path}">${part.text}</a>`;
        }).join(' / ');
        
        pathDisplay.innerHTML = `<i class="fas fa-folder-open"></i> ${breadcrumbHtml}`;
        
        // Insert before file list
        const fileListContainer = document.querySelector('.file-list');
        const existingPathDisplay = document.querySelector('.path-display');
        
        if (existingPathDisplay) {
            existingPathDisplay.replaceWith(pathDisplay);
        } else {
            fileListContainer.parentNode.insertBefore(pathDisplay, fileListContainer);
        }
        
        // Add event listeners to breadcrumb links
        document.querySelectorAll('.breadcrumb-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                this.loadFiles(link.dataset.path);
            });
        });
    }

    renderFiles(files) {
        this.elements.fileList.innerHTML = files.map(file => this.createFileElement(file)).join('');
        
        // Add event listeners to checkboxes
        document.querySelectorAll('.file-checkbox').forEach(checkbox => {
            checkbox.addEventListener('change', () => this.handleFileSelection(checkbox));
        });
        
        // Add event listeners to directory links
        document.querySelectorAll('.directory-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                this.loadFiles(link.dataset.path);
            });
        });
    }

    createFileElement(file) {
        const icon = file.isDirectory ? 'fa-folder' : this.getFileIcon(file.name);
        
        // Special case for parent directory
        if (file.isParentDir) {
            return `
                <div class="file-item">
                    <div class="checkbox-cell">
                        <!-- No checkbox for parent directory -->
                    </div>
                    <div class="name-cell">
                        <a href="#" class="directory-link" data-path="${file.path}">
                            <i class="fas fa-folder-up"></i>
                            ${file.name}
                        </a>
                    </div>
                    <div class="size-cell"></div>
                    <div class="date-cell"></div>
                </div>
            `;
        }
        
        // Directory with checkbox
        if (file.isDirectory) {
            return `
                <div class="file-item" data-path="${file.path}">
                    <div class="checkbox-cell">
                        <input type="checkbox" class="file-checkbox folder-checkbox" 
                               data-path="${file.path}" 
                               data-size="0" 
                               data-is-folder="true">
                    </div>
                    <div class="name-cell">
                        <a href="#" class="directory-link" data-path="${file.path}">
                            <i class="fas ${icon}"></i>
                            ${file.name}
                        </a>
                    </div>
                    <div class="size-cell">${this.formatSize(file.size)}</div>
                    <div class="date-cell">${this.formatDate(file.lastModified)}</div>
                </div>
            `;
        } else {
            return `
                <div class="file-item" data-path="${file.path}">
                    <div class="checkbox-cell">
                        <input type="checkbox" class="file-checkbox" 
                               data-path="${file.path}" 
                               data-size="${file.size}" 
                               data-is-folder="false">
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
    }

    handleFileSelection(checkbox) {
        const isFolder = checkbox.dataset.isFolder === 'true';
        const path = checkbox.dataset.path;
        
        if (checkbox.checked) {
            // Add to appropriate set
            if (isFolder) {
                this.selectedFolders.add(path);
            } else {
                this.selectedFiles.add(path);
            }
        } else {
            // Remove from appropriate set
            if (isFolder) {
                this.selectedFolders.delete(path);
            } else {
                this.selectedFiles.delete(path);
            }
        }

        this.updateUI();
    }

    updateUI() {
        const selectedFilesCount = this.selectedFiles.size;
        const selectedFoldersCount = this.selectedFolders.size;
        const totalSelectedCount = selectedFilesCount + selectedFoldersCount;
        
        this.elements.downloadBtn.disabled = totalSelectedCount === 0;
        
        // Update selection count display
        let selectionText = '';
        if (selectedFilesCount > 0 && selectedFoldersCount > 0) {
            selectionText = `${selectedFilesCount} file${selectedFilesCount !== 1 ? 's' : ''}, ${selectedFoldersCount} folder${selectedFoldersCount !== 1 ? 's' : ''} selected`;
        } else if (selectedFilesCount > 0) {
            selectionText = `${selectedFilesCount} file${selectedFilesCount !== 1 ? 's' : ''} selected`;
        } else if (selectedFoldersCount > 0) {
            selectionText = `${selectedFoldersCount} folder${selectedFoldersCount !== 1 ? 's' : ''} selected`;
        } else {
            selectionText = '0 items selected';
        }
        this.elements.selectedCount.textContent = selectionText;

        // Calculate total size (file sizes only, folder sizes would require an API call)
        const totalSize = Array.from(document.querySelectorAll('.file-checkbox:not([data-is-folder="true"]):checked'))
            .reduce((sum, checkbox) => sum + parseInt(checkbox.dataset.size), 0);
        this.elements.totalSize.textContent = `Total: ${this.formatSize(totalSize)}`;

        // Update select all checkbox
        const checkboxes = document.querySelectorAll('.file-checkbox');
        const checkedCount = document.querySelectorAll('.file-checkbox:checked').length;
        this.elements.selectAll.checked = checkedCount > 0 && checkedCount === checkboxes.length;
        this.elements.selectAll.indeterminate = checkedCount > 0 && checkedCount < checkboxes.length;
    }

    filterFiles(searchTerm) {
        const filtered = this.files.filter(file => 
            file.name.toLowerCase().includes(searchTerm.toLowerCase())
        );
        this.renderFiles(filtered);
    }

    async downloadSelectedFiles() {
        if (this.selectedFiles.size === 0 && this.selectedFolders.size === 0) return;

        // If we have selected folders, we need to fetch all files in those folders
        if (this.selectedFolders.size > 0) {
            try {
                // For each selected folder, get all files recursively
                for (const folderPath of this.selectedFolders) {
                    await this.addFolderFilesToSelection(folderPath);
                }
            } catch (error) {
                console.error('Error preparing folder download:', error);
                this.showError('Failed to prepare folder download');
                return;
            }
        }

        // Build the query string with all files to download
        const queryString = Array.from(this.selectedFiles)
            .map(file => `files=${encodeURIComponent(file)}`)
            .join('&');

        // Create a hidden link and trigger the download
        const downloadUrl = `${this.API_BASE_URL}/download?${queryString}`;
        window.location.href = downloadUrl;
    }

    // Recursively fetch files from a folder and add them to selectedFiles
    async addFolderFilesToSelection(folderPath) {
        try {
            const response = await fetch(`${this.API_BASE_URL}/files?path=${encodeURIComponent(folderPath)}`);
            const data = await response.json();
            const files = data.files || [];
            
            // Add all files from this folder to selectedFiles
            for (const file of files) {
                if (file.isDirectory) {
                    // Recursively add files from subfolders
                    await this.addFolderFilesToSelection(file.path);
                } else {
                    // Add individual file
                    this.selectedFiles.add(file.path);
                }
            }
        } catch (error) {
            console.error(`Error fetching files from folder ${folderPath}:`, error);
            throw error;
        }
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
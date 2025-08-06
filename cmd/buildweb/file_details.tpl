<!DOCTYPE html>
<html>
<head>
    <title>{{.Package.Package.Package}} - {{.BuiltFile.Builder}}/{{.BuiltFile.Target}} Archive Contents</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            margin: 0; 
            padding: 40px;
            background-color: #f8f9fa;
            color: #333;
        }
        .container { 
            max-width: 1200px; 
            margin: 0 auto; 
            background: white; 
            border-radius: 8px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1); 
            padding: 30px; 
        }
        h1 { color: #2c3e50; margin-bottom: 10px; font-size: 2.5em; }
        h2 { color: #34495e; border-bottom: 2px solid #e9ecef; padding-bottom: 10px; margin-top: 30px; }
        .back-link { 
            margin-bottom: 20px; 
            display: inline-block;
            padding: 8px 16px;
            background: #6c757d;
            color: white;
            border-radius: 5px;
            text-decoration: none;
            margin-right: 10px;
        }
        .back-link:hover { background: #545b62; color: white; }
        .info-section { margin: 25px 0; }
        .info-table { 
            border-collapse: collapse; 
            width: 100%; 
            background: white;
            border-radius: 6px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .info-table th, .info-table td { 
            padding: 12px 16px; 
            text-align: left; 
            border-bottom: 1px solid #e9ecef;
        }
        .info-table th { 
            background-color: #f8f9fa; 
            font-weight: 600;
            width: 180px; 
            color: #495057;
        }
        .info-table tr:last-child td { border-bottom: none; }
        
        .file-list {
            background: white;
            border-radius: 6px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            overflow: hidden;
            max-height: 600px;
            overflow-y: auto;
        }
        .file-list-header {
            background-color: #f8f9fa;
            padding: 12px 16px;
            border-bottom: 2px solid #e9ecef;
            display: flex;
            font-weight: 600;
            color: #495057;
            position: sticky;
            top: 0;
            z-index: 10;
        }
        .file-name-col { flex: 1; }
        .file-size-col { width: 100px; text-align: right; }
        
        .file-item {
            display: flex;
            align-items: center;
            padding: 10px 16px;
            border-bottom: 1px solid #e9ecef;
            min-height: 20px;
        }
        .file-item:last-child { border-bottom: none; }
        .file-item:nth-child(even) { background-color: #f8f9fa; }
        .file-item:hover { background-color: #e3f2fd; }
        
        .file-name {
            flex: 1;
            font-family: 'SF Mono', Consolas, monospace;
            font-size: 0.9em;
            word-break: break-all;
            color: #495057;
        }
        .file-size {
            width: 100px;
            text-align: right;
            font-family: 'SF Mono', Consolas, monospace;
            font-size: 0.85em;
            color: #6c757d;
        }
        
        .download-buttons {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
        }
        .download-btn {
            padding: 12px 24px;
            border-radius: 6px;
            text-decoration: none;
            text-align: center;
            font-weight: 500;
            transition: all 0.2s;
        }
        .download-btn-primary {
            background: #007bff;
            color: white;
        }
        .download-btn-primary:hover {
            background: #0056b3;
            color: white;
        }
        .download-btn-secondary {
            background: #6c757d;
            color: white;
        }
        .download-btn-secondary:hover {
            background: #545b62;
            color: white;
        }
        
        .stats {
            background: #e9ecef;
            padding: 15px;
            border-radius: 6px;
            margin-bottom: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .stats-left { font-weight: 500; }
        .stats-right { 
            font-family: 'SF Mono', Consolas, monospace;
            font-size: 0.9em;
            color: #495057;
        }
        
        code { 
            background: #f8f9fa; 
            padding: 2px 4px; 
            border-radius: 3px; 
            font-family: 'SF Mono', Consolas, monospace;
            font-size: 0.9em;
        }
        
        .builder-badge {
            display: inline-block;
            background: #e9ecef;
            color: #495057;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.85em;
            font-weight: 500;
            margin-left: 10px;
        }
        
        .target-badge {
            display: inline-block;
            background: #d4edda;
            color: #155724;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.85em;
            font-weight: 500;
            margin-left: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <a href="../../../{{.Package.Package.Package}}.html" class="back-link">← Back to Package</a>
        <a href="../../../builder_{{.BuiltFile.Builder}}.html" class="back-link">← Back to {{.BuiltFile.Builder}}</a>
        
        <h1>{{.Package.Package.Package}}</h1>
        <div style="margin-bottom: 20px;">
            <span class="builder-badge">{{.BuiltFile.Builder}}</span>
            <span class="target-badge">{{.BuiltFile.Target}}</span>
        </div>

        <div class="download-buttons">
            <a href="../../../{{getBuiltFilePath .Package.Package.Package .BuiltFile.ArchPath}}" class="download-btn download-btn-primary" download>
                ⬇ Download Archive ({{formatFileSize .BuiltFile.FileSize}})
            </a>
            <a href="../../../{{getBuiltFilePath .Package.Package.Package .BuiltFile.InfoPath}}" class="download-btn download-btn-secondary" target="_blank">
                📄 Build Info
            </a>
        </div>

        <div class="info-section">
            <h2>Build Information</h2>
            <table class="info-table">
                <tr><th>Package</th><td>{{.Package.Package.Package}}</td></tr>
                <tr><th>Version</th><td>{{.Package.Package.Version}}</td></tr>
                <tr><th>Builder</th><td>{{.BuiltFile.Builder}}</td></tr>
                <tr><th>Target</th><td>{{.BuiltFile.Target}}</td></tr>
                <tr><th>Build ID</th><td><code>{{.BuiltFile.ID}}</code></td></tr>
                <tr><th>Archive Size</th><td>{{formatFileSize .BuiltFile.FileSize}}</td></tr>
            </table>
        </div>

        <div class="info-section">
            <h2>Archive Contents</h2>
            {{$archiveInfo := getArchiveInfo .BuiltFile.ArchPath}}
            {{if $archiveInfo.Files}}
            <div class="stats">
                <div class="stats-left">{{len $archiveInfo.Files}} files in archive</div>
                <div class="stats-right">
                    Total uncompressed size: {{formatFileSize $archiveInfo.TotalSize}}
                </div>
            </div>
            <div class="file-list">
                <div class="file-list-header">
                    <div class="file-name-col">File Path</div>
                    <div class="file-size-col">Size</div>
                </div>
                {{range $archiveInfo.Files}}
                <div class="file-item">
                    <div class="file-name">{{.Name}}</div>
                    <div class="file-size">{{formatFileSize .Size}}</div>
                </div>
                {{end}}
            </div>
            {{else}}
            <div class="stats">
                <div class="stats-left">Unable to read archive contents</div>
                <div class="stats-right">Archive may be corrupted or use an unsupported format</div>
            </div>
            {{end}}
        </div>
    </div>
</body>
</html>
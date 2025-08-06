<!DOCTYPE html>
<html>
<head>
    <title>Package Information</title>
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
        .stats { color: #6c757d; font-size: 1.1em; margin-bottom: 30px; }
        table { 
            border-collapse: collapse; 
            width: 100%; 
            background: white;
            border-radius: 6px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        th, td { 
            padding: 12px 16px; 
            text-align: left; 
            border-bottom: 1px solid #e9ecef;
        }
        th { 
            background-color: #f8f9fa; 
            font-weight: 600;
            color: #495057;
        }
        tr:last-child td { border-bottom: none; }
        tr:nth-child(even) { background-color: #f8f9fa; }
        .type-badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.8em;
            font-weight: 500;
        }
        .type-host { background: #e3f2fd; color: #1976d2; }
        .type-native { background: #f3e5f5; color: #7b1fa2; }
        a { 
            color: #007bff; 
            text-decoration: none; 
            padding: 6px 12px;
            border-radius: 4px;
            transition: background-color 0.2s;
        }
        a:hover { 
            background-color: #e3f2fd;
            text-decoration: none; 
        }
        .progress-bar {
            width: 100px;
            height: 20px;
            background-color: #e9ecef;
            border-radius: 10px;
            overflow: hidden;
            display: inline-block;
            vertical-align: middle;
            margin-left: 10px;
            box-shadow: inset 0 1px 3px rgba(0,0,0,0.2);
        }
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #28a745 0%, #20c997 100%);
            border-radius: 10px;
            transition: width 0.3s ease;
            position: relative;
        }
        .progress-text {
            position: absolute;
            width: 100px;
            text-align: center;
            line-height: 20px;
            font-size: 11px;
            font-weight: 600;
            color: #333;
            text-shadow: 1px 1px 1px rgba(255,255,255,0.8);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Package Information</h1>
        <p class="stats">Total packages: {{len .}}</p>
        
        <div style="margin-bottom: 30px;">
            <h2>Builder Matrices</h2>
            <div style="display: flex; gap: 10px; flex-wrap: wrap;">
                {{range getBuilders}}
                <a href="builder_{{.}}.html" style="background: #007bff; color: white; padding: 8px 16px; border-radius: 5px; text-decoration: none; font-weight: 500;">{{.}}</a>
                {{end}}
            </div>
        </div>
        <table>
            <tr>
                <th>Package</th>
                <th>Version</th>
                <th>Type</th>
                <th>Build Progress</th>
                <th>Details</th>
            </tr>
            {{range .}}
            <tr>
                <td><strong>{{.Package.Package}}</strong></td>
                <td>{{.Package.Version}}</td>
                <td>
                    <span class="type-badge type-{{.Package.Type}}">{{.Package.Type}}</span>
                </td>
                <td>
                    {{$progress := getBuildProgress .}}
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: {{$progress}}%"></div>
                        <div class="progress-text">{{$progress}}%</div>
                    </div>
                </td>
                <td><a href="{{.Package.Package}}.html">View Details →</a></td>
            </tr>
            {{end}}
        </table>
    </div>
</body>
</html>`
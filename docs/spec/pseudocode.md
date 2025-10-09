# Pseudocode Design

## 1. Main Application Flow

```pseudocode
PROGRAM Jot

INITIALIZE:
    config = LoadConfiguration()
    logger = SetupLogging()

MAIN:
    command = ParseCLIArguments()

    SWITCH command:
        CASE "init":
            InitializeProject()
        CASE "build":
            BuildDocumentation()
        CASE "serve":
            StartWebServer()
        CASE "watch":
            WatchAndRebuild()
        CASE "export":
            ExportForLLM()
        DEFAULT:
            ShowHelp()
```

## 2. File Scanning Algorithm

```pseudocode
FUNCTION ScanMarkdownFiles(rootPath, ignorePatterns):
    documents = []
    ignorer = CreateIgnoreFilter(ignorePatterns)

    FUNCTION WalkDirectory(path, relativePath):
        entries = ReadDirectory(path)

        FOR EACH entry IN entries:
            fullPath = JoinPath(path, entry.name)
            relPath = JoinPath(relativePath, entry.name)

            IF ignorer.ShouldIgnore(relPath):
                CONTINUE

            IF entry.IsDirectory():
                WalkDirectory(fullPath, relPath)
            ELSE IF entry.HasExtension(".md"):
                doc = CreateDocument(fullPath, relPath)
                documents.APPEND(doc)

    WalkDirectory(rootPath, "")
    RETURN documents

FUNCTION CreateDocument(filePath, relativePath):
    content = ReadFile(filePath)
    metadata = ExtractFrontmatter(content)
    title = ExtractTitle(content, metadata)

    RETURN Document{
        Path: filePath,
        RelativePath: relativePath,
        Content: content,
        Title: title,
        Metadata: metadata,
        ModTime: GetFileModTime(filePath)
    }
```

## 3. TOC Generation Algorithm

```pseudocode
FUNCTION GenerateTableOfContents(documents):
    root = TOCNode{Title: "Root", Children: []}

    FOR EACH doc IN documents:
        pathParts = SplitPath(doc.RelativePath)
        currentNode = root

        FOR i = 0 TO LENGTH(pathParts) - 1:
            part = pathParts[i]
            isFile = (i == LENGTH(pathParts) - 1)

            child = FindChildByName(currentNode, part)

            IF child == NULL:
                child = TOCNode{
                    ID: GenerateID(pathParts[0:i+1]),
                    Title: IF isFile THEN doc.Title ELSE Humanize(part),
                    Path: IF isFile THEN doc.RelativePath ELSE NULL,
                    Children: []
                }
                currentNode.Children.APPEND(child)

            currentNode = child

    RETURN ConvertToXML(root)

FUNCTION ConvertToXML(node):
    xml = "<toc version='1.0'>\n"

    FUNCTION ProcessNode(n, depth):
        indent = REPEAT("  ", depth)

        IF n.Path != NULL:
            xml += indent + "<chapter id='" + n.ID + "' path='" + n.Path + "'>\n"
        ELSE:
            xml += indent + "<section id='" + n.ID + "'>\n"

        xml += indent + "  <title>" + EscapeXML(n.Title) + "</title>\n"

        FOR EACH child IN n.Children:
            ProcessNode(child, depth + 1)

        xml += indent + (IF n.Path != NULL THEN "</chapter>\n" ELSE "</section>\n")

    FOR EACH child IN node.Children:
        ProcessNode(child, 1)

    xml += "</toc>"
    RETURN xml
```

## 4. HTML Compilation Algorithm

```pseudocode
FUNCTION CompileToHTML(documents, tocXML, outputPath):
    toc = ParseXML(tocXML)
    linkMap = BuildLinkMap(documents)

    FOR EACH doc IN documents:
        html = ConvertMarkdownToHTML(doc.Content)
        html = ResolveInternalLinks(html, linkMap)
        html = AddSyntaxHighlighting(html)

        pageData = {
            Title: doc.Title,
            Content: html,
            TOC: toc,
            Breadcrumb: GenerateBreadcrumb(doc.RelativePath),
            Navigation: GenerateNavigation(doc, documents)
        }

        finalHTML = RenderTemplate("page.html", pageData)
        outputFile = JoinPath(outputPath, ChangeExtension(doc.RelativePath, ".html"))

        CreateDirectories(GetDirectory(outputFile))
        WriteFile(outputFile, finalHTML)

    CopyAssets(outputPath)
    GenerateSearchIndex(documents, outputPath)

FUNCTION ResolveInternalLinks(html, linkMap):
    PATTERN = /\[([^\]]+)\]\(([^)]+\.md[^)]*)\)/g

    RETURN html.REPLACE(PATTERN, FUNCTION(match, text, link):
        cleanLink = RemoveFragment(link)

        IF linkMap.HAS(cleanLink):
            htmlLink = ChangeExtension(linkMap[cleanLink], ".html")
            RETURN "[" + text + "](" + htmlLink + ")"
        ELSE:
            RETURN match
    )
```

## 5. Search Implementation

```pseudocode
FUNCTION BuildSearchIndex(documents):
    index = {
        documents: [],
        terms: {},
        fuzzy: FuzzyMatcher{}
    }

    FOR EACH doc IN documents:
        docIndex = LENGTH(index.documents)

        index.documents.APPEND({
            id: doc.RelativePath,
            title: doc.Title,
            path: ChangeExtension(doc.RelativePath, ".html")
        })

        // Extract and index terms
        terms = ExtractTerms(doc.Content + " " + doc.Title)

        FOR EACH term IN terms:
            normalizedTerm = Normalize(term)

            IF NOT index.terms.HAS(normalizedTerm):
                index.terms[normalizedTerm] = []

            index.terms[normalizedTerm].APPEND({
                doc: docIndex,
                count: CountOccurrences(doc.Content, term),
                positions: FindPositions(doc.Content, term)
            })

            index.fuzzy.Add(normalizedTerm)

    RETURN index

FUNCTION Search(query, index):
    results = []
    queryTerms = ExtractTerms(query)

    FOR EACH term IN queryTerms:
        normalized = Normalize(term)

        // Exact matches
        IF index.terms.HAS(normalized):
            FOR EACH occurrence IN index.terms[normalized]:
                AddToResults(results, occurrence)

        // Fuzzy matches
        fuzzyMatches = index.fuzzy.Find(normalized, maxDistance: 2)
        FOR EACH match IN fuzzyMatches:
            IF index.terms.HAS(match):
                FOR EACH occurrence IN index.terms[match]:
                    AddToResults(results, occurrence, fuzzyPenalty: 0.8)

    // Sort by relevance
    results.SORT(BY relevance DESCENDING)

    RETURN results.TAKE(maxResults: 20)
```

## 6. LLM Export Algorithm

```pseudocode
FUNCTION ExportForLLM(documents, format):
    SWITCH format:
        CASE "json":
            RETURN ExportJSON(documents)
        CASE "yaml":
            RETURN ExportYAML(documents)
        CASE "embeddings":
            RETURN GenerateEmbeddings(documents)

FUNCTION ExportJSON(documents):
    export = {
        version: "1.0",
        generated: CurrentTimestamp(),
        project: LoadProjectMetadata(),
        documents: []
    }

    FOR EACH doc IN documents:
        sections = ExtractSections(doc.Content)
        codeBlocks = ExtractCodeBlocks(doc.Content)
        links = ExtractLinks(doc.Content)

        export.documents.APPEND({
            id: GenerateDocID(doc.RelativePath),
            path: doc.RelativePath,
            title: doc.Title,
            content: doc.Content,
            html: ConvertToHTML(doc.Content),
            sections: sections,
            metadata: doc.Metadata,
            links: links,
            codeBlocks: codeBlocks
        })

    export.index = BuildSemanticIndex(export.documents)

    RETURN JSONStringify(export, pretty: true)

FUNCTION GenerateEmbeddings(documents):
    embeddings = []

    FOR EACH doc IN documents:
        chunks = SplitIntoChunks(doc.Content, maxTokens: 512, overlap: 128)

        FOR EACH chunk IN chunks:
            embedding = {
                documentId: doc.RelativePath,
                chunkId: GenerateChunkID(),
                text: chunk.text,
                vector: NULL,  // Placeholder for actual embedding
                metadata: {
                    title: doc.Title,
                    section: chunk.section,
                    startLine: chunk.startLine,
                    endLine: chunk.endLine
                }
            }

            embeddings.APPEND(embedding)

    RETURN embeddings
```

## 7. Version Control Integration

```pseudocode
FUNCTION InitializeVersionControl():
    vcs = {
        store: CreateFileStore(".jot-history"),
        currentVersion: 0
    }

    RETURN vcs

FUNCTION TrackChanges(documents, vcs):
    changes = []

    FOR EACH doc IN documents:
        previousVersion = vcs.GetPrevious(doc.RelativePath)

        IF previousVersion == NULL:
            changes.APPEND({
                type: "created",
                path: doc.RelativePath,
                content: doc.Content
            })
        ELSE IF doc.Content != previousVersion.Content:
            changes.APPEND({
                type: "modified",
                path: doc.RelativePath,
                diff: GenerateDiff(previousVersion.Content, doc.Content)
            })

    // Check for deletions
    previousPaths = vcs.GetAllPaths()
    currentPaths = documents.MAP(d => d.RelativePath)

    FOR EACH path IN previousPaths:
        IF NOT currentPaths.CONTAINS(path):
            changes.APPEND({
                type: "deleted",
                path: path
            })

    IF LENGTH(changes) > 0:
        vcs.Commit(changes)
        BroadcastChanges(changes)

    RETURN changes
```

## 8. Web Server Implementation

```pseudocode
FUNCTION StartWebServer(config):
    server = CreateGinServer()

    // Static file serving
    server.Static("/", config.outputPath)

    // API endpoints
    server.GET("/api/docs", ListDocuments)
    server.GET("/api/docs/:id", GetDocument)
    server.GET("/api/search", SearchDocuments)
    server.GET("/api/toc", GetTableOfContents)
    server.GET("/api/export/:format", ExportDocumentation)

    // WebSocket for hot reload
    IF config.autoReload:
        server.GET("/ws", WebSocketHandler)
        StartFileWatcher(config.inputPaths, FUNCTION(change):
            BroadcastReload(change)
        )

    server.Run(":" + config.port)

FUNCTION WebSocketHandler(connection):
    clients.ADD(connection)

    connection.ON("close", FUNCTION():
        clients.REMOVE(connection)
    )

    // Keep alive
    EVERY 30 seconds:
        connection.Send({type: "ping"})
```

## 9. CLI Command Handlers

```pseudocode
FUNCTION HandleBuildCommand(flags):
    config = MergeConfigs(LoadConfigFile(), flags)

    ShowProgress("Scanning files...")
    documents = ScanMarkdownFiles(config.inputPaths, config.ignorePatterns)

    ShowProgress("Generating TOC...")
    toc = GenerateTableOfContents(documents)

    ShowProgress("Compiling HTML...")
    CompileToHTML(documents, toc, config.outputPath)

    IF config.features.search:
        ShowProgress("Building search index...")
        index = BuildSearchIndex(documents)
        WriteFile(JoinPath(config.outputPath, "search-index.json"), index)

    IF config.features.versioning:
        ShowProgress("Tracking changes...")
        TrackChanges(documents, vcs)

    ShowSuccess("Build complete! " + LENGTH(documents) + " files processed.")

FUNCTION HandleWatchCommand(flags):
    config = MergeConfigs(LoadConfigFile(), flags)

    // Initial build
    HandleBuildCommand(flags)

    // Watch for changes
    watcher = CreateFileWatcher(config.inputPaths)

    watcher.ON("change", FUNCTION(event):
        ShowInfo("Change detected: " + event.path)

        // Incremental rebuild
        IF event.type == "modify" OR event.type == "create":
            doc = CreateDocument(event.path, GetRelativePath(event.path))
            UpdateDocument(doc)
        ELSE IF event.type == "delete":
            RemoveDocument(event.path)

        RebuildAffectedFiles()
    )

    ShowInfo("Watching for changes... Press Ctrl+C to stop.")
    watcher.Start()
```

## 10. Error Handling

```pseudocode
FUNCTION SafeExecute(operation, errorMessage):
    TRY:
        RETURN operation()
    CATCH error:
        LogError(error)

        IF IsFileNotFound(error):
            ShowError(errorMessage + ": File not found - " + error.path)
            ShowHint("Check if the file exists and you have read permissions")
        ELSE IF IsPermissionError(error):
            ShowError(errorMessage + ": Permission denied - " + error.path)
            ShowHint("Check file permissions or run with appropriate privileges")
        ELSE IF IsParseError(error):
            ShowError(errorMessage + ": Parse error - " + error.details)
            ShowHint("Check markdown syntax at line " + error.line)
        ELSE:
            ShowError(errorMessage + ": " + error.message)

        IF config.debug:
            PrintStackTrace(error)

        EXIT(1)
```

This pseudocode provides a comprehensive design for all major components of the Jot documentation generator, ready for implementation in Go.

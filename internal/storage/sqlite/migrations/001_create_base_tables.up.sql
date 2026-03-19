CREATE TABLE "snippets" (
	"id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"command" TEXT NOT NULL,
	"description" TEXT,
	"created_at" TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
	"updated_at" TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
	"use_count" INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE "tags" (
	"id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"name" TEXT NOT NULL UNIQUE
);

CREATE TABLE "snippet_tags" (
	"snippet_id" INTEGER NOT NULL,
	"tag_id" INTEGER NOT NULL,
	PRIMARY KEY ("snippet_id", "tag_id"),
	FOREIGN KEY ("snippet_id") REFERENCES "snippets"("id") ON DELETE CASCADE,
	FOREIGN KEY ("tag_id") REFERENCES "tags"("id") ON DELETE CASCADE
);

CREATE INDEX "idx_tags_name" ON "tags"("name");
CREATE INDEX "idx_snippet_tags_snippet_id" ON "snippet_tags"("snippet_id");
CREATE INDEX "idx_snippet_tags_tag_id" ON "snippet_tags"("tag_id");

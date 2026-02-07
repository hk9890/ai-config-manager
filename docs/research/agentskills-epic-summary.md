# Epic: AgentSkills.in Marketplace Integration

**Epic ID**: ai-config-manager-660  
**Priority**: P1  
**Status**: Ready to start  
**Phase**: Research & Architecture Design

---

## Overview

Integrate agentskills.in marketplace as a **source provider** for aimgr, enabling users to discover and import skills from their 100,000+ skill catalog.

### Strategic Position

agentskills.in is **complementary**, not a competitor:
- **They provide**: Discovery layer (marketplace, search, 100k+ skills)
- **We provide**: Local repository management, project integration, multi-source sync

### Target User Workflow

```yaml
# ~/.config/aimgr/aimgr.yaml
sync:
  sources:
    - url: ~/dev/my-skills              # Local filesystem
    - url: gh:anthropics/skills         # GitHub
    - url: agentskills:@anthropic/xlsx  # AgentSkills marketplace (NEW)
```

```bash
# Update from all sources
aimgr repo sync

# Use in project
cd my-project
aimgr install skill/xlsx
```

---

## Research Phase Tasks

**All tasks are P1 priority and ready to start:**

### 1. Analyze Source Provider Architecture (ai-config-manager-032)

**Questions:**
- Do we have clear interfaces/abstractions for source providers?
- How does GitHub source work? (parse → resolve → discover → import)
- How does local source work?
- Where should agentskills fit?

**Output:** `docs/research/source-provider-architecture.md`
- Current architecture diagram
- Source provider comparison table
- Interface analysis
- Recommendations for agentskills integration

### 2. Workspace Caching Strategy (ai-config-manager-9nf)

**Questions:**
- Does agentskills.in need workspace caching?
- API returns full content - no Git clone needed
- Should we cache API responses? For how long?
- What's the update flow?

**Scenarios to evaluate:**
- **A**: No workspace (fetch → temp → import → delete)
- **B**: Workspace for skills (cache downloaded content)
- **C**: Hybrid (use GitHub URL from API → workspace)

**Output:** Section in `docs/research/source-provider-architecture.md`
- Workspace benefits analysis
- Recommendation with rationale
- Implementation sketch

### 3. Metadata Tracking (ai-config-manager-4ag)

**Questions:**
- What metadata do we currently track?
- What agentskills.in metadata should we track?
  - skill ID, author, scopedName, githubUrl, stars, updatedAt
- How to handle scoped names (@author/skill)?
- How to handle name conflicts?
- How to check for updates?

**Output:** Section in `docs/research/source-provider-architecture.md`
- Current metadata tracking approach
- Proposed schema for agentskills sources
- Scoped name handling decision
- Update detection mechanism

### 4. API Client Design (ai-config-manager-6v1)

**Questions:**
- What API endpoints do we need?
- Error handling strategy (network, 404, rate limits)?
- Content acquisition priority (content field → rawUrl → githubUrl)?
- Caching strategy (HTTP, in-memory, disk)?

**Output:** Section in `docs/research/source-provider-architecture.md`
- Client interface design
- Error handling strategy
- Content acquisition flow diagram
- Caching recommendations

---

## Acceptance Gate (ai-config-manager-2h5)

**Blocked by:** Epic + all 4 research tasks

**Criteria:**
- [ ] **Functional**: Can import skills: `aimgr repo import agentskills:@anthropic/xlsx`
- [ ] **Functional**: Sync config works with agentskills sources
- [ ] **Functional**: Imported skills install correctly
- [ ] **Architectural**: Source provider abstraction exists
- [ ] **Architectural**: Workspace strategy decided and implemented
- [ ] **Architectural**: Metadata tracking implemented
- [ ] **Architectural**: Error handling robust
- [ ] **Documentation**: Architecture documented
- [ ] **Documentation**: README updated
- [ ] **Documentation**: Code documented
- [ ] **Quality**: Tests passing
- [ ] **Quality**: Code reviewed
- [ ] **Quality**: Performance acceptable

**Owner**: beads-verify-agent

---

## Non-Goals (Future Epics)

These are explicitly **OUT OF SCOPE** for this epic:

- ❌ Search command (`aimgr search agentskills python`)
- ❌ Interactive skill selection UI
- ❌ Category browsing
- ❌ Lock file / update checking
- ❌ Telemetry

**This epic focuses solely on backend integration**: getting agentskills.in working as a source provider in a clean, extensible way.

---

## Key Architecture Questions

1. **Do we have clear source provider interfaces?**
   - Need to investigate existing code structure
   - Is there a Source interface/trait?
   - Can agentskills follow same pattern as GitHub?

2. **How does GitHub source work?**
   - Entry point: Where is `gh:owner/repo` parsed?
   - Resolution: How converted to local path?
   - Workspace: How does it interact with `pkg/workspace`?
   - Can we follow the same flow?

3. **Is workspace caching useful for agentskills?**
   - GitHub: Clones Git repos to workspace (10-50x faster)
   - agentskills: API provides full content (no clone needed)
   - Should we cache API responses? Or always fetch fresh?

4. **How to track metadata?**
   - GitHub: Do we store source URL?
   - agentskills: Need skill ID, author, for update checking
   - Where is metadata stored? (in resource files? separate?)

5. **How to handle scoped names?**
   - agentskills: `@author/name` format
   - aimgr: `skill/name` format (no namespacing)
   - Options:
     - A: `skills/pdf-processing/` (drop author)
     - B: `skills/@anthropic/pdf-processing/` (preserve)
     - C: `skills/pdf-processing/` + metadata tracks author

---

## Success Metrics

**Phase 1 (Research) Complete When:**
- All 4 research tasks closed
- `docs/research/source-provider-architecture.md` exists
- Architecture questions answered
- Clear path forward for implementation

**Epic Complete When:**
- Acceptance gate passes
- Can import from agentskills.in via CLI and config
- Backend architecture clean and extensible
- Documentation complete

---

## Next Steps

1. **Start Research Tasks** (all ready to work, no blockers)
   - Can be done in parallel or sequentially
   - Recommended order:
     1. Source provider architecture (foundation)
     2. Workspace caching (affects implementation)
     3. Metadata tracking (affects schema)
     4. API client design (implementation details)

2. **Review Research Findings**
   - Discuss architecture decisions
   - Validate approach
   - Adjust plan if needed

3. **Implementation Tasks** (will be created after research)
   - Based on research findings
   - Will block acceptance gate

---

## References

- **Competitor Analysis**: `docs/research/competitor-analysis-agent-skills-cli.md`
- **Integration Plan**: `docs/research/agentskills-integration-plan.md`
- **agentskills.in API**: https://www.agentskills.in/api/skills
- **API Documentation**: https://www.agentskills.in/docs (if exists)

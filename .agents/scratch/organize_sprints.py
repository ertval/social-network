import re
import os
import glob

sprint_files = sorted(glob.glob('/home/ertval/code/zone-modules/social-network/docs/plan/sprints/sprint-*.md'))
tracker_file = '/home/ertval/code/zone-modules/social-network/docs/plan/sprints/ticket-tracker.md'

# 5-developer assignees mapping
assignees_map = {
    # Sprint 0
    "S0-BE-01": "BE-A",
    "S0-BE-02": "BE-B",
    "S0-BE-03": "SD-QA",
    "S0-BE-04": "BE-A",
    "S0-BE-05": "BE-B",
    "S0-FE-01": "FE-A",
    "S0-FE-02": "FE-B",
    "S0-DEV-01": "SD-QA",
    "S0-DEV-02": "SD-QA",
    "S0-DEV-03": "SD-QA",
    
    # Sprint 1
    "S1-BE-01": "BE-A",
    "S1-BE-02": "BE-B",
    "S1-BE-03": "BE-B",
    "S1-BE-04": "BE-A",
    "S1-BE-05": "BE-A",
    "S1-BE-06": "BE-B",
    "S1-BE-07": "BE-A",
    "S1-BE-08": "BE-B",
    "S1-BE-09": "BE-B",
    "S1-BE-10": "BE-A",
    "S1-BE-11": "SD-QA",
    "S1-FE-01": "FE-A",
    "S1-FE-02": "FE-A",
    "S1-FE-03": "FE-B",
    "S1-FE-04": "SD-QA",
    
    # Sprint 2
    "S2-BE-01": "BE-A",
    "S2-BE-02": "BE-A",
    "S2-BE-03": "BE-A",
    "S2-BE-04": "BE-A",
    "S2-BE-05": "BE-A",
    "S2-BE-06": "BE-A",
    "S2-BE-07": "BE-A",
    "S2-BE-08": "BE-A",
    "S2-BE-09": "BE-A",
    "S2-BE-10": "BE-A",
    "S2-BE-11": "BE-A",
    "S2-BE-12": "SD-QA",
    "S2-BE-13": "BE-B",
    "S2-BE-14": "BE-B",
    "S2-BE-15": "BE-B",
    "S2-BE-16": "BE-B",
    "S2-BE-17": "BE-B",
    "S2-BE-18": "BE-B",
    "S2-BE-19": "BE-B",
    "S2-BE-20": "BE-B",
    "S2-BE-21": "BE-B",
    "S2-BE-22": "SD-QA",
    "S2-FE-01": "FE-A",
    "S2-FE-02": "FE-A",
    "S2-FE-03": "FE-A",
    "S2-FE-04": "FE-A",
    "S2-FE-05": "FE-B",
    "S2-FE-06": "FE-B",
    "S2-FE-07": "FE-B",
    "S2-FE-08": "SD-QA",
    
    # Sprint 3
    "S3-BE-01": "BE-A",
    "S3-BE-02": "BE-A",
    "S3-BE-03": "BE-A",
    "S3-BE-04": "BE-A",
    "S3-BE-05": "BE-A",
    "S3-BE-06": "BE-A",
    "S3-BE-07": "BE-A",
    "S3-BE-08": "BE-A",
    "S3-BE-09": "BE-A",
    "S3-BE-10": "BE-A",
    "S3-BE-11": "BE-A",
    "S3-BE-12": "SD-QA",
    "S3-BE-13": "BE-B",
    "S3-BE-14": "BE-B",
    "S3-BE-15": "BE-B",
    "S3-BE-16": "BE-B",
    "S3-BE-17": "BE-B",
    "S3-BE-18": "SD-QA",
    "S3-BE-19": "BE-B",
    "S3-BE-20": "BE-B",
    "S3-BE-21": "BE-B",
    "S3-BE-22": "BE-B",
    "S3-BE-23": "BE-B",
    "S3-BE-24": "BE-B",
    "S3-FE-01": "FE-A",
    "S3-FE-02": "FE-A",
    "S3-FE-03": "FE-A",
    "S3-FE-04": "FE-B",
    "S3-FE-05": "FE-B",
    "S3-FE-06": "FE-B",
    "S3-FE-07": "SD-QA",
    "S3-FE-08": "SD-QA",
    
    # Sprint 4
    "S4-BE-01": "BE-A",
    "S4-BE-02": "BE-A",
    "S4-BE-03": "BE-A",
    "S4-BE-04": "BE-A",
    "S4-BE-05": "BE-A",
    "S4-BE-06": "BE-A",
    "S4-BE-07": "BE-A",
    "S4-BE-08": "BE-A",
    "S4-BE-09": "BE-A",
    "S4-BE-10": "BE-A",
    "S4-BE-11": "BE-A",
    "S4-BE-12": "BE-A",
    "S4-BE-13": "BE-A",
    "S4-BE-14": "BE-A",
    "S4-BE-15": "BE-A",
    "S4-BE-16": "BE-B",
    "S4-BE-17": "BE-B",
    "S4-BE-18": "BE-B",
    "S4-BE-19": "BE-B",
    "S4-BE-20": "BE-B",
    "S4-BE-21": "BE-B",
    "S4-FE-01": "FE-A",
    "S4-FE-02": "FE-A",
    "S4-FE-03": "FE-A",
    "S4-FE-04": "FE-A",
    "S4-FE-05": "FE-B",
    "S4-FE-06": "FE-B",
    "S4-FE-07": "FE-B",
    "S4-FE-08": "SD-QA",
    
    # Sprint 5
    "S5-BE-01": "BE-A",
    "S5-BE-02": "BE-A",
    "S5-BE-03": "BE-A",
    "S5-BE-04": "BE-A",
    "S5-BE-05": "BE-A",
    "S5-BE-06": "BE-A",
    "S5-BE-07": "BE-A",
    "S5-BE-08": "SD-QA",
    "S5-BE-09": "BE-B",
    "S5-BE-10": "BE-B",
    "S5-BE-11": "BE-B",
    "S5-BE-12": "BE-B",
    "S5-BE-13": "BE-B",
    "S5-BE-14": "BE-B",
    "S5-BE-15": "BE-B",
    "S5-BE-16": "SD-QA",
    "S5-FE-01": "FE-A",
    "S5-FE-02": "FE-A",
    "S5-FE-03": "FE-A",
    "S5-FE-04": "FE-B",
    "S5-FE-05": "FE-B",
    "S5-FE-06": "SD-QA",
    "S5-FE-07": "SD-QA",
    
    # Sprint 6
    "S6-BE-01": "BE-A",
    "S6-BE-02": "BE-A",
    "S6-BE-03": "BE-B",
    "S6-BE-04": "BE-A + BE-B",
    "S6-BE-05": "SD-QA",
    "S6-BE-06": "SD-QA",
    "S6-BE-07": "SD-QA",
    "S6-BE-08": "SD-QA",
    "S6-FE-01": "SD-QA",
    "S6-FE-02": "FE-A",
    "S6-FE-03": "FE-B",
    "S6-FE-04": "SD-QA",
    "S6-FE-05": "SD-QA",
    "S6-FE-06": "FE-A + FE-B",
    "S6-FE-07": "SD-QA",
    "S6-DEV-01": "SD-QA",
    "S6-DEV-02": "SD-QA",
    "S6-DEV-03": "SD-QA",
    "S6-DEV-04": "SD-QA",
    "S6-DEV-05": "SD-QA"
}

groups_order = [
    'BE-A',
    'BE-B',
    'BE-A + BE-B',
    'SD-QA',
    'FE-A',
    'FE-B',
    'FE-A + FE-B'
]

group_headers = {
    'BE-A': 'BE-A (Backend A)',
    'BE-B': 'BE-B (Backend B)',
    'BE-A + BE-B': 'Joint BE-A & BE-B',
    'SD-QA': 'SD-QA (System Design/QA)',
    'FE-A': 'FE-A (Frontend A)',
    'FE-B': 'FE-B (Frontend B)',
    'FE-A + FE-B': 'Joint FE-A & FE-B'
}

sprint_names = {
    0: "Sprint 0: Foundation (Week 1–2)",
    1: "Sprint 1: Platform & Core Infrastructure (Week 3–4)",
    2: "Sprint 2: User & Topic Features (Week 5–6)",
    3: "Sprint 3: Follow, Comment & Notification (Week 7–8)",
    4: "Sprint 4: Group & Event Features (Week 9–10)",
    5: "Sprint 5: Chat & OAuth (Week 11–12)",
    6: "Sprint 6: Integration, Cleanup & Polish (Week 13–14)"
}

all_tickets_registry = {}  # ticket_id -> {assignee, name, sprint}

def parse_sprint_file(content):
    match = re.search(r'\n###\s+', content)
    if not match:
        return content, []
    
    first_ticket_index = match.start() + 1
    preamble = content[:first_ticket_index]
    tickets_content = content[first_ticket_index:]
    
    # Clean preamble: keep up to '---'
    pre_lines = preamble.strip().split('\n')
    cleaned_pre_lines = []
    found_divider = False
    for line in pre_lines:
        cleaned_pre_lines.append(line)
        if line.strip() == '---':
            found_divider = True
            break
    
    if found_divider:
        preamble = '\n'.join(cleaned_pre_lines) + '\n\n'
    else:
        while pre_lines and (pre_lines[-1].strip().startswith('##') or not pre_lines[-1].strip()):
            pre_lines.pop()
        preamble = '\n'.join(pre_lines) + '\n\n'
    
    # Split the tickets_content by '###'
    ticket_blocks = []
    parts = re.split(r'\n(###\s+)', '\n' + tickets_content)
    i = 1
    while i < len(parts):
        header = parts[i]
        body = parts[i+1]
        
        body_lines = body.split('\n')
        # Filter out lines starting with '##' from the body
        body_lines = [line for line in body_lines if not line.strip().startswith('##')]
        while body_lines and (body_lines[-1].strip() == '---' or not body_lines[-1].strip()):
            body_lines.pop()
        body_cleaned = '\n'.join(body_lines)
        
        ticket_text = header + body_cleaned
        
        # Extract ticket ID and Name
        id_name_match = re.search(r'###\s+((S\d-\w+-\d+):?\s*(.*))', ticket_text)
        ticket_id = ""
        ticket_name = ""
        if id_name_match:
            ticket_id = id_name_match.group(2).strip()
            ticket_name = id_name_match.group(3).strip()
        
        # Override assignee based on 5-dev mapping
        if ticket_id in assignees_map:
            assignee = assignees_map[ticket_id]
            # Replace the assignee line in the ticket text
            ticket_text = re.sub(r'\*\s+\*\*Assignee:\*\*\s*.*', f'* **Assignee:** {assignee}', ticket_text)
        else:
            assignee_match = re.search(r'\*\s+\*\*Assignee:\*\*\s*(.*)', ticket_text)
            assignee = assignee_match.group(1).strip() if assignee_match else "UNKNOWN"
        
        ticket_blocks.append({
            'id': ticket_id,
            'name': ticket_name,
            'text': ticket_text,
            'assignee': assignee
        })
        i += 2
        
    return preamble, ticket_blocks

# 1. Process and load sprint files (they are already reorganized by assignee on disk)
for fpath in sprint_files:
    fname = os.path.basename(fpath)
    sprint_num = int(re.search(r'sprint-(\d+)', fname).group(1))
    
    with open(fpath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    preamble, tickets = parse_sprint_file(content)
    
    for t in tickets:
        if t['id']:
            all_tickets_registry[t['id']] = {
                'assignee': t['assignee'],
                'name': t['name'],
                'sprint': sprint_num
            }

# 2. Process and rewrite ticket-tracker.md: Group by Sprint, then by Assignee
with open(tracker_file, 'r', encoding='utf-8') as f:
    tracker_content = f.read()

# Get preamble up to the first '---'
match_divider = re.search(r'\n---\n', tracker_content)
if match_divider:
    tracker_preamble = tracker_content[:match_divider.end()]
else:
    tracker_preamble = tracker_content

# Build new ticket tracker grouped by Sprint, then by Assignee
sprint_sections = []
for s_num in sorted(sprint_names.keys()):
    sect_header = f"## {sprint_names[s_num]}"
    sect_content = f"{sect_header}\n"
    
    # Get all tickets in this sprint
    s_tickets_ids = [tid for tid, info in all_tickets_registry.items() if info['sprint'] == s_num]
    
    # Group by assignee within the sprint
    for g in groups_order:
        g_s_tickets = [tid for tid in s_tickets_ids if all_tickets_registry[tid]['assignee'] == g]
        if g_s_tickets:
            # Sort tickets by ID
            g_s_tickets = sorted(g_s_tickets, key=lambda x: [int(c) if c.isdigit() else c for c in re.split(r'(\d+)', x)])
            sect_content += f"### {group_headers[g]}\n"
            for tid in g_s_tickets:
                name = all_tickets_registry[tid]['name']
                sect_content += f"- [ ] **{tid}:** {name}\n"
            sect_content += "\n"
            
    sprint_sections.append(sect_content.strip())

new_tracker_content = tracker_preamble + "\n\n" + "\n\n---\n\n".join(sprint_sections) + "\n"

with open(tracker_file, 'w', encoding='utf-8') as f:
    f.write(new_tracker_content)
print("Reorganized ticket-tracker.md to group by Sprint then Assignee")

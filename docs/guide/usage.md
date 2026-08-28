# Dashboards and access

## Create your first dashboard

Sign in, create a dashboard, and choose a unique URL slug. The slug becomes part of its address, such as `/infrastructure`. Add groups to organize related services, then add links inside each group.

Use edit mode to drag groups into a new order or move links within and between groups. Layout changes are saved to the server.

<div class="screenshot-placeholder">
  <div><strong>Dashboard editor placeholder</strong><br>Add a screenshot showing groups, items, and edit controls.</div>
</div>

## Visibility

Choose visibility based on who should reach the dashboard:

| Visibility | Anonymous visitor | Signed-in user | Owner or admin |
| --- | ---: | ---: | ---: |
| Public | ✓ | ✓ | ✓ |
| Authenticated users | — | ✓ | ✓ |
| Private | — | — | ✓ |

Administrators can select a non-private dashboard as the system main dashboard. This is the landing view for users who do not have a personal default and for anonymous visitors when it is public.

An individual user can mark one of their own dashboards as their personal default.

## Layout and appearance

- **Rows** places groups in horizontal sections.
- **Columns** organizes groups into columns.
- **Masonry** packs groups based on their rendered height.
- **Wide mode** uses more of the browser width.
- **Clean mode** reduces surrounding navigation for a focused display.

Icons may use Iconify names, uploaded files, or remote image URLs. Separate dark-mode icons can be supplied where needed.

## Availability checks

Enable ping on a link to display whether its destination is reachable. “Only when down” keeps healthy links quiet and surfaces the indicator only when a check fails.

Availability is a convenience signal from the DashLit server, not a replacement for full monitoring and alerting.

## Import, export, and cloning

Dashboard export creates a portable DashLit JSON file containing dashboard settings, groups, and items. Import it while signed in to create a new owned dashboard. Slug conflicts are resolved automatically.

You can also clone complete dashboards, individual groups, or individual items when building similar views.

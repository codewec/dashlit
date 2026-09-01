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

## Icon search and theme variants

The icon picker searches two sources independently:

- **selfh.st/icons** focuses on software and services commonly used in self-hosted environments.
- **Iconify** provides a much broader collection of general-purpose icon sets.

Iconify results are displayed as soon as they arrive. The selfh.st search continues independently and adds its results without blocking or replacing Iconify results.

For selfh.st icons, DashLit prefers SVG and falls back to PNG when SVG is unavailable. Selecting a selfh.st icon for the Default field automatically fills the empty Dark field with the corresponding theme variant. An existing manually selected Dark icon is never overwritten.

The selfh.st convention uses a dark-colored asset on light backgrounds and a light-colored asset on dark backgrounds. When no separate theme asset exists, DashLit uses the standard icon for both fields. Downloaded selfh.st assets are served through the DashLit API and cached on disk.

Many monochrome Iconify assets are black by default. DashLit renders Iconify icons in a light color on dark application themes so they remain visible. This adjustment applies only to Iconify; branded selfh.st icons retain their supplied colors and theme variants.

The picker shows Default icons on a light preview background and Dark icons on a dark preview background. You can still override either field manually, upload a file, or enter a remote image URL.

## Availability checks

Enable the availability check on a link to display whether its destination is reachable. By default DashLit checks the link URL, but you can provide a separate check URL—for example, an internal address or a dedicated health endpoint. Enable **Skip TLS verification** only when that endpoint uses a trusted self-signed or mismatched certificate. “Only when down” keeps healthy links quiet and surfaces the indicator only when a check fails.

Availability is a convenience signal from the DashLit server, not a replacement for full monitoring and alerting.

## Import, export, and cloning

Dashboard export creates a portable DashLit JSON file containing dashboard settings, groups, and items. Import it while signed in to create a new owned dashboard. Slug conflicts are resolved automatically.

You can also clone complete dashboards, individual groups, or individual items when building similar views. From a group's actions menu, choose **Copy to dashboard…** to select another dashboard you can edit and copy the group there together with all of its links and settings.

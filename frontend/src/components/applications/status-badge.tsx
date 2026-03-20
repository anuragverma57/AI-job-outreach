import { Badge } from "@/components/ui/badge";
import type { ApplicationStatus } from "@/types/application";

const statusConfig: Record<
  ApplicationStatus,
  { label: string; variant: "default" | "secondary" | "destructive" | "outline" }
> = {
  draft: { label: "Draft", variant: "secondary" },
  applied: { label: "Applied", variant: "default" },
  replied: { label: "Replied", variant: "outline" },
  interview: { label: "Interview", variant: "outline" },
  offer: { label: "Offer", variant: "default" },
  rejected: { label: "Rejected", variant: "destructive" },
  ghosted: { label: "Ghosted", variant: "secondary" },
};

export function StatusBadge({ status }: { status: ApplicationStatus }) {
  const config = statusConfig[status] ?? statusConfig.draft;
  return <Badge variant={config.variant}>{config.label}</Badge>;
}

export function mapTaskFromApi(item = {}) {
  const status = item.status === 'published' ? 'available' : item.status
  const assigneeName = item.currentAssignee?.username || ''
  const createdById = item.createdBy || ''
  const publishedById = item.publishedBy || ''

  return {
    id: item.id,
    title: item.title,
    summary: item.descriptionPlain || '',
    descriptionHtml: item.descriptionHtml || '',
    reward: item.bounty ?? 0,
    deadline: item.deadline,
    priority: item.priority || 'medium',
    tags: Array.isArray(item.tags) ? item.tags.slice() : [],
    status,
    assignee: assigneeName,
    assigneeId: item.currentAssignee?.userId || '',
    assigneeStatus: item.currentAssignee?.status || '',
    completedAt: item.currentAssignee?.completedAt || null,
    createdAt: item.createdAt,
    updatedAt: item.updatedAt,
    createdById,
    publishedById,
    ownerId: publishedById || createdById
  }
}

export type PrincipalId = 'vinay' | 'maya' | 'arjun' | 'paybot'
export type ScopeType = 'employee_self' | 'employee' | 'department' | 'tenant'
export type Permission = string

export interface Principal {
  id: PrincipalId
  name: string
  initials: string
  employeeId: string | null
  departmentId: string | null
  title: string
  color: string
}

export interface Scope {
  type: ScopeType
  targetId?: string
}

export interface Grant {
  id: number
  principalId: PrincipalId
  permission: Permission
  scope: Scope
  valid: boolean
}

export interface Resource {
  id: string
  label: string
  employeeId: string
  employeeName: string
  departmentId: string
  tenantId: string
  amount: string
}

export interface Challenge {
  id: number
  title: string
  brief: string
  principalId: PrincipalId
  permission: Permission
  resourceId: string
  expected: boolean
  points: number
}

export interface ScenarioPack {
  id: string
  name: string
  tenantId: string
  principals: Principal[]
  permissions: { value: Permission; label: string; verb: string }[]
  resources: Resource[]
  challenges: Challenge[]
  initialGrants: Grant[]
}

export const payrollScenario: ScenarioPack = {
  id: 'hrms-payroll',
  name: 'Payroll authority',
  tenantId: 'TENANT-001',
  principals: [
    { id: 'vinay', name: 'Vinay', initials: 'VS', employeeId: 'EMP-005', departmentId: 'ENG', title: 'Employee', color: '#51d6c3' },
    { id: 'maya', name: 'Maya', initials: 'MR', employeeId: 'EMP-018', departmentId: 'FIN', title: 'Payroll administrator', color: '#ffbe5c' },
    { id: 'arjun', name: 'Arjun', initials: 'AK', employeeId: 'EMP-023', departmentId: 'ENG', title: 'Engineering manager', color: '#a78bfa' },
    { id: 'paybot', name: 'PayBot', initials: 'AI', employeeId: null, departmentId: null, title: 'Delegated payroll agent', color: '#ff6f7d' },
  ],
  permissions: [
    { value: 'hrms:payroll:ledger::read', label: 'Read payroll ledger', verb: 'read' },
    { value: 'hrms:payroll:salary_earning::create', label: 'Create salary earning', verb: 'create' },
    { value: 'hrms:payroll:ledger::post', label: 'Post payroll ledger', verb: 'post' },
  ],
  resources: [
    { id: 'PAY-000005', label: 'September payroll', employeeId: 'EMP-005', employeeName: 'Vinay', departmentId: 'ENG', tenantId: 'TENANT-001', amount: '₹1,42,800' },
    { id: 'PAY-000018', label: 'September payroll', employeeId: 'EMP-018', employeeName: 'Maya', departmentId: 'FIN', tenantId: 'TENANT-001', amount: '₹1,85,200' },
    { id: 'PAY-000023', label: 'September payroll', employeeId: 'EMP-023', employeeName: 'Arjun', departmentId: 'ENG', tenantId: 'TENANT-001', amount: '₹1,67,500' },
    { id: 'PAY-EXT-007', label: 'External tenant payroll', employeeId: 'EMP-007', employeeName: 'Neha', departmentId: 'OPS', tenantId: 'TENANT-002', amount: '₹98,400' },
  ],
  challenges: [
    { id: 1, title: 'My salary, my eyes', brief: 'Vinay reads his own September ledger.', principalId: 'vinay', permission: 'hrms:payroll:ledger::read', resourceId: 'PAY-000005', expected: true, points: 100 },
    { id: 2, title: 'Curious colleague', brief: 'Vinay tries to read Arjun’s salary.', principalId: 'vinay', permission: 'hrms:payroll:ledger::read', resourceId: 'PAY-000023', expected: false, points: 100 },
    { id: 3, title: 'Payroll desk', brief: 'Maya reads an employee ledger in her tenant.', principalId: 'maya', permission: 'hrms:payroll:ledger::read', resourceId: 'PAY-000005', expected: true, points: 100 },
    { id: 4, title: 'Tenant escape', brief: 'Maya aims her tenant grant at an external ledger.', principalId: 'maya', permission: 'hrms:payroll:ledger::read', resourceId: 'PAY-EXT-007', expected: false, points: 150 },
    { id: 5, title: 'Naked mutation', brief: 'An employee attempts to post committed payroll.', principalId: 'vinay', permission: 'hrms:payroll:ledger::post', resourceId: 'PAY-000005', expected: false, points: 125 },
    { id: 6, title: 'Confused deputy', brief: 'A service agent acts without delegated authority.', principalId: 'paybot', permission: 'hrms:payroll:ledger::read', resourceId: 'PAY-000005', expected: false, points: 175 },
  ],
  initialGrants: [
    { id: 1, principalId: 'vinay', permission: 'hrms:payroll:ledger::read', scope: { type: 'employee_self' }, valid: true },
    { id: 2, principalId: 'maya', permission: 'hrms:payroll:ledger::read', scope: { type: 'tenant', targetId: 'TENANT-001' }, valid: true },
  ],
}

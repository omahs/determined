from argparse import ONE_OR_MORE, Namespace
from typing import Any, List

from determined.cli import render, setup_session
from determined.common.api import authentication, bindings
from determined.common.declarative_argparse import Arg, Cmd


@authentication.required
def get_resource_pools(args: Namespace) -> None:
    print("getting resource pools")
    resp = bindings.get_GetResourcePools(setup_session(args))

    headers = [
        "name",
        "default",
        "accelerator",
        "type",
        "instance type",
        "slots used",
        "agents",
        "scheduler",
    ]

    render.tabulate_or_csv(
        headers=headers,
        values=[
            [
                pool.name,
                "default compute"
                if pool.defaultComputePool
                else "default aux"
                if pool.defaultAuxPool
                else None,
                pool.accelerator,
                pool.type,
                pool.instanceType,
                str(pool.slotsUsed) + "/" + str(pool.slotsAvailable + pool.slotsUsed),
                str(pool.numAgents)
                + "/"
                + (
                    str(pool.numAgents)
                    if pool.type == bindings.v1ResourcePoolType.STATIC
                    else str(pool.maxAgents)
                ),
                pool.schedulerType,
            ]
            for pool in resp.resourcePools
        ],
        as_csv=False,
    )
    return


@authentication.required
def add_binding(args: Namespace) -> None:
    body = bindings.v1BindRPToWorkspaceRequest(
        resourcePoolName=args.pool_name, workspaceNames=args.workspace_names
    )
    bindings.post_BindRPToWorkspace(setup_session(args), body=body, resourcePoolName=args.pool_name)

    print(
        f'added bindings between the resource pool "{args.pool_name}" '
        f"and the following workspaces: {args.workspace_names}"
    )
    return


@authentication.required
def remove_binding(args: Namespace) -> None:
    body = bindings.v1UnbindRPFromWorkspaceRequest(
        resourcePoolName=args.pool_name,
        workspaceNames=args.workspace_names,
    )
    bindings.delete_UnbindRPFromWorkspace(
        setup_session(args), body=body, resourcePoolName=args.pool_name
    )

    print(
        f'removed bindings between the resource pool "{args.pool_name}" '
        f"and the following workspaces: {args.workspace_names}"
    )
    return


@authentication.required
def replace_bindings(args: Namespace) -> None:
    body = bindings.v1OverwriteRPWorkspaceBindingsRequest(
        resourcePoolName=args.pool_name,
        workspaceNames=args.workspace_names,
    )
    bindings.put_OverwriteRPWorkspaceBindings(
        setup_session(args), body=body, resourcePoolName=args.pool_name
    )

    print(
        f'replaced bindings of the resource pool "{args.pool_name}" '
        f"with those to the following workspaces: {args.workspace_names}"
    )
    return


@authentication.required
def list_workspaces(args: Namespace) -> None:
    session = setup_session(args)
    resp = bindings.get_ListWorkspacesBoundToRP(session, resourcePoolName=args.pool_name)
    workspace_names = ""

    if resp.workspaceIds:
        workspace_names = ", ".join(
            [
                workspace.name
                for workspace in bindings.get_GetWorkspaces(session).workspaces
                if workspace.id in set(resp.workspaceIds)
            ]
        )

    render.tabulate_or_csv(
        headers=["resource pool", "workspaces"],
        values=[[args.pool_name, workspace_names]],
        as_csv=False,
    )
    return


args_description = [
    Cmd(
        "resource-pool rp",
        None,
        "manage resource pools",
        [
            Cmd(
                "list ls",
                get_resource_pools,
                "list all resource_pools",
                [],
            ),
            Cmd(
                "bindings",
                None,
                "manage resource pool bindings",
                [
                    Cmd(
                        "add",
                        add_binding,
                        "add a resource-pool-to-workspace binding",
                        [
                            Arg("pool_name", type=str, help="name of the resource pool to bind"),
                            Arg(
                                "workspace_names",
                                nargs=ONE_OR_MORE,
                                type=str,
                                default=None,
                                help="the workspace to bind to",
                            ),
                        ],
                    ),
                    Cmd(
                        "remove",
                        remove_binding,
                        "remove a resource-pool-to-workspace binding",
                        [
                            Arg("pool_name", type=str, help="name of the resource pool to unbind"),
                            Arg(
                                "workspace_names",
                                nargs=ONE_OR_MORE,
                                type=str,
                                default=None,
                                help="the workspace to unbind from",
                            ),
                        ],
                    ),
                    Cmd(
                        "replace",
                        replace_bindings,
                        "replace all existing resource-pool-to-workspace bindings",
                        [
                            Arg("pool_name", type=str, help="name of the resource pool to bind"),
                            Arg(
                                "workspace_names",
                                nargs=ONE_OR_MORE,
                                type=str,
                                default=None,
                                help="the workspaces to bind to",
                            ),
                        ],
                    ),
                    Cmd(
                        "list-workspaces",
                        list_workspaces,
                        "list all workspaces bound to the pool",
                        [
                            Arg("pool_name", type=str, help="name of the resource pool"),
                        ],
                    ),
                ],
            ),
        ],
    )
]  # type: List[Any]

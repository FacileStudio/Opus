import { describe, expect, it } from "vitest";
import { stringifyIds } from "./stringify-ids";

describe("stringifyIds", () => {
  it("turns numeric id fields into strings at any depth", () => {
    expect(
      stringifyIds({
        data: {
          id: 7,
          workspaceId: 2,
          columns: [
            { id: "todo", tasks: [{ id: 12, projectId: 7, position: 3 }] },
          ],
        },
      }),
    ).toEqual({
      data: {
        id: "7",
        workspaceId: "2",
        columns: [
          { id: "todo", tasks: [{ id: "12", projectId: "7", position: 3 }] },
        ],
      },
    });
  });

  it("handles id arrays and leaves other numbers alone", () => {
    expect(stringifyIds({ taskIds: [1, 2], number: 4, total: 9 })).toEqual({
      taskIds: ["1", "2"],
      number: 4,
      total: 9,
    });
  });
});

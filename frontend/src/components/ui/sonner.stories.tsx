import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { toast } from "sonner";
import { Toaster } from "@/components/ui/sonner";

// More on how to set up stories at: https://storybook.js.org/docs/react/writing-stories/introduction#default-export
const meta = {
  title: "UI/Sonner",
  component: Toaster,
  parameters: {
    // Optional parameter to center the component in the Canvas. More info: https://storybook.js.org/docs/react/configure/story-layout
    layout: "centered",
  },
  // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/react/writing-docs/autodocs
  tags: ["autodocs"],
} satisfies Meta<typeof Toaster>;

export default meta;
type Story = StoryObj<typeof meta>;

// More on writing stories with args: https://storybook.js.org/docs/react/writing-stories/args
export const Default: Story = {
  render: (args) => {
    const showToast = () => {
      toast("Event has been created", {
        description: "Friday, February 10, 2023 at 5:57 PM",
      });
    };

    return (
      <div>
        <button onClick={showToast}>Show Toast</button>
        <Toaster {...args} />
      </div>
    );
  },
};

export const WithAction: Story = {
  render: (args) => {
    const showToast = () => {
      toast("Uh oh! Something went wrong.", {
        description: "There was a problem with your request.",
        action: {
          label: "Try again",
          onClick: () => console.log("Try again clicked"),
        },
      });
    };

    return (
      <div>
        <button onClick={showToast}>Show Toast with Action</button>
        <Toaster {...args} />
      </div>
    );
  },
};

export const Destructive: Story = {
  render: (args) => {
    const showToast = () => {
      toast.error("Uh oh! Something went wrong.", {
        description: "There was a problem with your request.",
      });
    };

    return (
      <div>
        <button onClick={showToast}>Show Destructive Toast</button>
        <Toaster {...args} />
      </div>
    );
  },
};

export const Success: Story = {
  render: (args) => {
    const showToast = () => {
      toast.success("Success!", {
        description: "Your action was completed successfully.",
      });
    };

    return (
      <div>
        <button onClick={showToast}>Show Success Toast</button>
        <Toaster {...args} />
      </div>
    );
  },
};

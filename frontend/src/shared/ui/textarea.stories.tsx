import type { Meta, StoryObj } from '@storybook/nextjs-vite';
import { Textarea } from '@/shared/ui/textarea';

// More on how to set up stories at: https://storybook.js.org/docs/react/writing-stories/introduction#default-export
const meta = {
  title: 'UI/Textarea',
  component: Textarea,
  parameters: {
    // Optional parameter to center the Canvas. More info: https://storybook.js.org/docs/react/configure/story-layout
    layout: 'centered',
  },
  // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/react/writing-docs/autodocs
  tags: ['autodocs'],
} satisfies Meta<typeof Textarea>;

export default meta;
type Story = StoryObj<typeof meta>;

// More on writing stories with args: https://storybook.js.org/docs/react/writing-stories/args
export const Default: Story = {
  args: {
    placeholder: 'Enter your message...',
  },
};

export const WithLabel: Story = {
  render: (args) => (
    <div className="grid w-full gap-1.5">
      <label htmlFor="message">Your message</label>
      <Textarea
        id="message"
        placeholder="Type your message here..."
        {...args}
      />
    </div>
  ),
};

export const Disabled: Story = {
  args: {
    disabled: true,
    placeholder: 'Disabled textarea...',
  },
};

export const WithValue: Story = {
  args: {
    defaultValue: 'This is a sample text in the textarea.',
  },
};

export const Large: Story = {
  args: {
    placeholder: 'Enter a longer message...',
    className: 'h-32',
  },
};
